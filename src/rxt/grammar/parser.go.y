%{
package grammar

import (
  "errors"

  "fmt"
  "strings"
  "github.com/simelo/rextporter/src/config"
  "github.com/simelo/rextporter/src/memconfig"
  // "github.com/davecgh/go-spew/spew"
  "os"
  // "github.com/simelo/rextporter/src/util"
)

var (
  ErrBlockLevelUnderflow = errors.New("End of block is not possible beyond DATABASE")
  ErrBlockLevelOverflow  = errors.New("Too many nested syntax levels")
)

type strTuple struct {
  key  string
  val  string
}

func newStrTuple(s1, s2 string) *strTuple {
  return &strTuple {
    key:  s1,
    val: s2,
  }
}

// FIXME : Not global. Parser stack ? TLS ?
var root config.RextRoot
var service config.RextServiceDef
func init() {
  // TODO(denisacostaq@gmail.com): services
  root = memconfig.NewRootConfig(nil)
  service = &memconfig.Service{}
}

%}

%union{
  root config.RextRoot
  service config.RextServiceDef
  services []config.RextServiceDef
  resource config.RextResourceDef
  resources []config.RextResourceDef
  decoder config.RextDecoderDef
  nodeSolver config.RextNodeSolver
  metric config.RextMetricDef
  metrics []config.RextMetricDef
  label config.RextLabelDef
  auth config.RextAuthDef
  options config.RextKeyValueStore
  key     string
  strval  string
  strlist []string
  pair    *strTuple
  identVal int
}

%start dataset

%token <key> GET POST COUNTER GAUGE HISTOGRAM SUMMARY
%token <strval> IDENTIFIER STR_LITERAL RESOURCE_PATH
%token FROM HELP LABELS METRIC NAME SET TYPE TO DESCRIPTION WITH_OPTIONS AS
%token BIE BLK EOB EOL CTX EOF COMMA DEFINE_AUTH EXTRACT_USING FOR_SERVICE FOR_STACK DATASET


// defverb srcverb mtvalue id setcls optsblk strlst idlst mlabels mname mtype mhelp optional_metric_fieldso optional_metric_fields metsec metblk extblk sseco ssec srcsecsufixo srcsecsufix srcsec auth_optso defsec optblko srvstk srvstko srvcls srvclso mainsec mainblk eolo 
//         eolo 
%type <key> defverb srcverb mtvalue mtype
%type <strval> id mname mhelp
// %type <pair> 
%type <options> optblko setcls optsblk auth_optso
%type <strlist> strlst idlst mlabels
%type <metric> optional_metric_fields metsec
%type <metrics> optional_metric_fieldso metblk
%type <resource> mainsec extblk sseco srcsec srcsecsufixo srcsecsufix
%type <auth> defsec
%type <services> srvstk srvstko // FIXME(denisacostaq@gmail.com): remove this
%type <services> srvcls srvclso
%type <service> mainblk

%%

defverb : DEFINE_AUTH
          { $$ = "AUTH" }

srcverb : GET
          { $$ = $1 }
        | POST
          { $$ = $1 }

mtvalue : GAUGE
          { $$ = config.KeyMetricTypeGauge }
        | COUNTER
          { $$ = config.KeyMetricTypeCounter }
        | HISTOGRAM
          { $$ = config.KeyMetricTypeHistogram }
        | SUMMARY
          { $$ = config.KeyMetricTypeSummary }

id : IDENTIFIER
     { $$ = $1 }

setcls : SET STR_LITERAL TO STR_LITERAL
         { 
           $$ = memconfig.NewOptionsMap()
            // TODO(denisacostaq@gmail.com): handle errors
            _, _ = $$.SetString($2, $4)
         }

optsblk : setcls
          {
            $$ = memconfig.NewOptionsMap()
            for _, k := range $1.GetKeys() {
              // TODO(denisacostaq@gmail.com): handle errors
              v, _ := $1.GetObject(k)
              // TODO(denisacostaq@gmail.com): handle errors
              _, _ = $$.SetObject(k, v)
            }
          }
        | optsblk EOL setcls
          {
            // TODO(denisacostaq@gmail.com): handle errors and previously exist
            $$, _ = memconfig.MergeStoresInANewOne($1, $3)
          }
        | /* empty */
          { $$ = nil }

strlst : STR_LITERAL
         { $$ = []string{ $1 } }
       | strlst COMMA STR_LITERAL
         { $$ = append($1, $3) }

idlst : id
        { $$ = []string{ $1 } }
      | idlst COMMA id
        { $$ = append($1, $3) }

mlabels : LABELS strlst
          { $$ = $2 }

mname : NAME STR_LITERAL
        { $$ = $2 }

mtype : TYPE mtvalue
        { $$ = $2 }

mhelp : HELP STR_LITERAL
        { $$ = $2 }

optional_metric_fieldso : /* empty */
                         { $$ = nil }
                       | optional_metric_fieldso EOL optional_metric_fields
                         { $$ = append($$, $3) }

optional_metric_fields : mlabels
                         {
                           $$ = &memconfig.MetricDef{}
                           for _, l := range $1 {
                             label := &memconfig.LabelDef{}
                             label.SetName(l)
                             $$.AddLabel(label)
                           }
                          }
                       | mhelp
                         { 
                           $$ = &memconfig.MetricDef{}
                           $$.SetMetricDescription($1)
                         }
                       | WITH_OPTIONS BLK optsblk EOB
                         {
                           $$ = &memconfig.MetricDef{}
                           opts := $$.GetOptions()
                           for _, k := range $3.GetKeys() {
                             // TODO(denisacostaq@gmail.com): handle error
                             v, _ := $3.GetObject(k)
                             opts.SetObject(k, v)
                           }
                         }

metsec : METRIC BLK mname EOL mtype optional_metric_fieldso EOB
          {
            var mtrDescription string
            var mtrLabels []config.RextLabelDef
            var mtrOptions, labelOpts config.RextKeyValueStore
            const prex = "\"label_path:"
            for _, optinalField := range $6 {
              if len(optinalField.GetMetricDescription()) != 0 {
                mtrDescription = optinalField.GetMetricDescription()
              }
              if len(optinalField.GetLabels()) != 0 {
                labelOpts = memconfig.NewOptionsMap()
                if mtrLabels == nil {
                  mtrLabels = optinalField.GetLabels()
                } else {
                  // FIXME(denisacostaq@gmail.com): error multiple labels definitions
                }
              }
              if opts := optinalField.GetOptions(); len(opts.GetKeys()) != 0 {
                  // TODO(denisacostaq@gmail.com): handle error
                  mtrOptions = memconfig.NewOptionsMap()
                  for _, k := range opts.GetKeys() {
                    // TODO(denisacostaq@gmail.com): handle errors
                    v, _ := opts.GetString(k)
                    if strings.HasPrefix(k, prex) {
                      // fmt.Println("dddddddddddddddddddddddddddddddddddddddddddddd", k)
                      if labelOpts != nil {
                        labelOpts.SetString(k, v)
                      } else {
                        // FIXME(denisacostaq@gmail.com):  handle error not labels defined
                      }
                    } else {
                      mtrOptions.SetString(k, v)
                    }
                  }
              }
              // spew.Dump("TTTTTTTTTTTTTT", labelOpts)
              if len(optinalField.GetLabels()) > 0 && len(optinalField.GetLabels()) != len(labelOpts.GetKeys()) {
                // FIXME(denisacostaq@gmail.com):  handle error labels and paths size do not match
              }
            }
            if len(mtrDescription) == 0 {
              mtrDescription = "Metric extracted by [rextporter](https://github.com/simelo/rextporter)"
            }
            $$ = &memconfig.MetricDef{}
            $$.SetMetricName($3)
            $$.SetMetricType($5)
            $$.SetMetricDescription(mtrDescription)
            // spew.Dump("_____________", labelOpts)
            for _, l := range mtrLabels {
              // TODO(denisacostaq@gmail.com): handle errors
              str := prex + l.GetName()
              k := str[:len(prex)] + str[len(prex)+1:]
              lp, _ := labelOpts.GetString(k)
              var ns config.RextNodeSolver
              ns = &memconfig.NodeSolver{}
              ns.SetNodePath(lp)
              l.SetNodeSolver(ns)
              $$.AddLabel(l)
            }
            // spew.Dump("optsoptsoptsoptsopts", $$.GetLabels())
            if mtrOptions != nil {
              opts := $$.GetOptions()
              for _, k := range mtrOptions.GetKeys() {
                // TODO(denisacostaq@gmail.com): handle errors
                v, _ := mtrOptions.GetObject(k)
                // TODO(denisacostaq@gmail.com): handle errors and previously exist
                _, _ = opts.SetObject(k, v)
              }
            }
            fmt.Println("dddddddddddddddd")
            // spew.Dump($$.GetLabels())
          }

metblk : metsec
         { $$ = []config.RextMetricDef{$1} }
       |  metblk EOL metsec
         { $$ = append($1, $3) }

extblk : EXTRACT_USING STR_LITERAL BLK optblko metblk EOB
          {
            $$ = &memconfig.ResourceDef{}
            var opts config.RextKeyValueStore
            if $4 != nil {
              // TODO(denisacostaq@gmail.com): handle errors
              opts, _ = $4.Clone()
            }
            decoder := memconfig.NewDecoder($2, opts)
            $$.SetDecoder(decoder)
            for _, mtr := range $5 {
              $$.AddMetricDef(mtr)
            }
          }

sseco : /* empty */
        { $$ = nil }
      |
        EOL extblk
        { $$ = $2 }

srcsecsufixo : /* empty */
               { $$ = nil }
             |
               BLK srcsecsufix EOB
               {
                 $$ = $2
               }

srcsecsufix : optblko sseco
            {
              $$ = $2
              if $1 != nil {
                if $$ == nil {
                  $$ = &memconfig.ResourceDef{}
                }
                opts := $$.GetOptions()
                for _, k := range $1.GetKeys() {
                  // TODO(denisacostaq@gmail.com): handle errors
                  v, _ := $1.GetObject(k)
                  // TODO(denisacostaq@gmail.com): handle errors and previously exist
                  _, _ = opts.SetObject(k, v)
                }
              }
            }

srcsec : srcverb IDENTIFIER FROM RESOURCE_PATH srcsecsufixo
         {
           $$ = $5
           if $$ == nil {
             $$ = &memconfig.ResourceDef{}
           }
           $$.SetResourceURI($4)
           switch $2 {
             // TODO(denisacostaq@gmail.com): solve this
             case "rest_api", "forward_metrics":
             default:
               fmt.Println("invalid type", $2)
           }
           // TODO(denisacostaq@gmail.com): $1
         }

auth_optso : BLK optblko EOB
             { $$ = $2 }

defsec : defverb IDENTIFIER AS STR_LITERAL auth_optso
         {
        //   if $1 == "AUTH" {
        //     $$ = &memconfig.HTTPAuth{}
        //     if $2 == "rest_csrf" {
        //       $$.SetAuthType(config.AuthTypeCSRF)
        //     }
        //     if $5 != nil {
        //         opts := $$.GetOptions()
        //         for _, k := range $5.GetKeys() {
        //           // TODO(denisacostaq@gmail.com): handle errors
        //           v, _ := $5.GetObject(k)
        //           // TODO(denisacostaq@gmail.com): handle errors and previously exist
        //           _, _ = opts.SetObject(k, v)
        //         }
        //     }
        //   } else {
            $$ = nil
        //   }
         }

optblko : /* empty */
          { $$ = nil }
        | WITH_OPTIONS BLK optsblk EOB
          { $$ = $3 }

// stkcls : 'FOR STACK' idlst
//          { $$ = nil /*$2*/ }
// stkclso : /* empty */
//           { $$ = nil }
//         | stkcls EOL
//           { $$ = nil /*$1*/ }

srvstk : FOR_STACK idlst
         { $$ = nil/*$2*/ }
srvstko : /* empty */
          { $$ = nil }
        | srvstk EOL
          { $$ = nil/*$1*/ }

srvcls : FOR_SERVICE idlst
         { 
            for _, id := range $2 {
              srv := &memconfig.Service{}
              srv.SetBasePath(id)
              $$ = append($$, srv)
            } 
         }

srvclso : /* empty */
          { $$ = nil }
        | srvcls EOL
          { $$ = $1 }

mainsec : defsec
          {
            service.SetAuthForBaseURL($1)
            $$ = nil
          }
        | srcsec
          {
            $$ = $1
          }

mainblk : /* empty */
          { $$ = nil }
        | mainblk EOL mainsec
          {
            if $3 != nil {
              service.AddResource($3)
            }
            $$ = service
          }

eolo : /* empty */
     | EOL
// eol_or_eof : EOL | EOF
dataset : eolo DATASET BLK srvclso srvstko optblko mainblk EOB eolo
          {
            // spew.Dump($7)
            // fmt.Println(len($7.GetResources()))
            for _, srv := range $4 {
              // service.SetName(srv)
              fmt.Println(srv)
            }
            // env = $1
            // $$ = env.NewServiceScraper()
            // if $5 != nil {
            //   // TODO : Error handling
            //   _ = env.RegisterScraperForServices($5...)
            // }
            // if $6 != nil {
            //   // TODO : Error handling
            //   _ = env.RegisterScraperForServices($6...)
            // }
            // if $7 != nil {
            //   util.MergeStoresInplace($$.GetOptions(), $7)
            // }
            // for _, mainsec := range $8 {
            //   if mainsec.src != nil {
            //     $$.AddSource(mainsec.src)
            //   } else if mainsec.value != nil {
            //     if auth, isAuth := mainsec.value.(core.RextAuth); isAuth {
            //       $$.AddAuthStrategy(auth, mainsec.key)
            //     }
            //     // TODO : Error handling
            //   }
            //   // TODO : Error handling
            // }
          }
%%

func Parse() int {
  yyErrorVerbose = true
	filename := "/usr/share/gocode/src/github.com/simelo/rextporter/src/rxt/testdata/skyexample.rxt"
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
  lex := NewLexer(file)
  state := lexerState{}
  cl := NewCustomLexer(lex, &state)
	e := yyParse(cl)
	fmt.Println("lexer -- ", "Return code:", e)
  return e
}