%{
package main

import (
  "errors"

  "fmt"
  "github.com/simelo/rextporter/src/config"
  "github.com/simelo/rextporter/src/core"
  // "github.com/davecgh/go-spew/spew"
  // "github.com/simelo/rextporter/src/util"
)

var (
  ErrBlockLevelUnderflow = errors.New("End of block is not possible beyond DATABASE")
  ErrBlockLevelOverflow  = errors.New("Too many nested syntax levels")
)

type parserEnv struct {
  env       core.RextEnv
  scraper   core.RextServiceScraper 
}

type strTuple struct {
  key  string
  val  string
}

type mainSecTuple struct {
  src  core.RextDataSource
  key  string
  val  interface{}
}

type metricDef struct {
  mname string
  mtype string
  mdesc string
  mlbls []string
  opts  core.RextKeyValueStore
}

// FIXME : Not global. Parser stack ? TLS ?
var root    parserEnv
var metric  metricDef

// TODO: metricDef should implement core.RextMetricDef

func value_for_str(str string) string {
  // FIXME: Support string literals
  return str[1: len(str) - 1]
}

func newOption() core.RextKeyValueStore {
  return config.NewOptionsMap()
}

func newStrTuple(s1, s2 string) *strTuple {
  return &strTuple {
    key:  s1,
    val: s2,
  }
}

func newMainDef(key string, value interface{}) *mainSecTuple {
  return &mainSecTuple{
    src: nil,
    key: key,
    val: value,
  }
}

func newMainSrc(src core.RextDataSource) *mainSecTuple {
  return &mainSecTuple{
    src: src,
    key: "",
    val: nil,
  }
}

func getRootEnv() *parserEnv {
  return &root
}

func (m *metricDef) GetMetricName() string {
  return m.mname
}

func (m *metricDef) GetMetricType() string {
  return m.mtype
}

func (m *metricDef) GetMetricDescription() string {
  return m.mdesc
}

func (m *metricDef) GetMetricLabels() []string {
  return m.mlbls
}

func (m *metricDef) SetMetricName(name string) {
  m.mname = name
}

func (m *metricDef) SetMetricType(typeid string) {
  m.mtype = typeid
}

func (m *metricDef) SetMetricDescription(desc string) {
  m.mdesc = desc
}

func (m *metricDef) SetMetricLabels(labels []string) {
  m.mlbls = labels
}

func (m *metricDef) GetOptions() core.RextKeyValueStore {
  return nil
}

%}

%union{
  root    core.RextServiceScraper
  options core.RextKeyValueStore
  mains   []mainSecTuple
  mainsec *mainSecTuple
  exts    []core.RextMetricsExtractor
  extract core.RextMetricsExtractor
  metrics []core.RextMetricDef
  metric  core.RextMetricDef
  key     string
  strval  string
  strlist []string
  pair    *strTuple
  identVal int
}

%start dataset

%token <key> COUNTER GAUGE HISTOGRAM SUMMARY
%token <strval> COMMA
%token <strval> STR_LITERAL RESOURCE_PATH
%token AS
%token BIE BLK EOB EOL CTX
%token DATASET
%token DEFINE_AUTH
%token DESCRIPTION
%token EXTRACT_USING
%token FOR_SERVICE
%token FOR_STACK
%token FROM
%token HELP
%token GET
%token<strval> IDENTIFIER
%token LABELS
%token METRIC
%token NAME
%token POST
%token SET
%token TO
%token TYPE

%type <extract> extblk
%type <exts> sseco ssec srcsecsufix srcsecsufixo 
%type <key>     mtvalue srcverb GET POST defverb// mfname
%type <mains>   mainblk
%type <mainsec> defsec mainsec srcsec
%type <metric>  metsec
%type <metrics> metblk
%type <options> optsblk optblkr optblkl optblko
%type <pair>    setcls
// %type <root>    dataset
%type <strlist> strlst idlst mlabels mlablso // stkcls stkclso srvcls srvclso
%type <strval>  mname mtype mhelp mhelpo // id 
%type <strval> id
%type <strlist> srvcls srvclso srvstk srvstko


%%

defverb : DEFINE_AUTH
          { $$ = "AUTH" }
        ;
srcverb : GET
          { fmt.Println("0000000000000000000000000000000000000"); $$ = $1 }
        | POST
          { $$ = $1 }
        ;
mtvalue : GAUGE
          { $$ = config.KeyMetricTypeGauge }
        | COUNTER
          { $$ = config.KeyMetricTypeCounter }
        | HISTOGRAM
          { $$ = config.KeyMetricTypeHistogram }
        | SUMMARY
          { $$ = config.KeyMetricTypeSummary }
        ;
id      : IDENTIFIER
          { $$ = $1 }
        // | STR
        //   { $$ = value_for_str($1) }
//         ;
setcls  : SET STR_LITERAL TO STR_LITERAL
          { $$ = newStrTuple($2, $4); fmt.Println("oOo", $$); }
optsblk : setcls
          {
            $$ = newOption()
            // TODO: Error handling
            _, _ = $$.SetString($1.key, $1.val)
          }
        | optsblk EOL setcls
          {
            // TODO: Error handling
            _, _ = $1.SetString($3.key, $3.val)
            $$ = $1
          }
strlst : STR_LITERAL
          { $$ = []string{ $1 } }
        | strlst COMMA STR_LITERAL
          { $$ = append($1, $3) }
idlst   : id
          { $$ = []string{ $1 } }
        | idlst COMMA id
          { $$ = append($1, $3) }
mlabels : LABELS strlst
          { $$ = $2 }
mname   : NAME STR_LITERAL
          { $$ = $2 }
mtype   : TYPE mtvalue
          { $$ = $2 }
mhelp   : HELP STR_LITERAL
          { $$ = $2 }
mhelpo  : /* empty */
          {
            $$ = "Metric extracted by [rextporter](https://github.com/simelo/rextporter)"
          }
        | EOL mhelp
          { $$ = $2 }
mlablso : /* empty */
           { $$ = nil }
        // | EOL mlabels
        //   { $$ = $2 }
optblkl : /* empty */
          { $$ = nil }
        | optsblk
          { $$ = $1 }
optblkr : /* empty */
          { $$ = nil }
        | optsblk
          { fmt.Println("ppppppppppppp"); $$ = $1 }
metsec  : METRIC BLK mname EOL mtype mhelpo mlablso optblkl EOB
          {
            mm := metricDef{
              mname: $3,
              mtype: $5,
              mdesc: $6,
              mlbls: $7,
              opts:  $8,
            }
            fmt.Println(mm)
            $$ = nil
          }
metblk  : metsec
          { $$ = nil; /*[]metricDef{ $1 }*/ }
        | metblk EOL metsec
          { $$ = append($1, $3) }
extblk  : EXTRACT_USING STR_LITERAL BLK optblkr metblk EOB
          {
            // env := getRootEnv()
            $$ = nil //env.NewMetricsExtractor($2, $4, $5)
            // for _, md := range $6 {
            //   $$.AddMetricRule(&md)
            // }
          }
sseco : /* empty */
        { $$ = nil }
      |
        ssec
        {
          $$ = $1;
        }
ssec    : extblk
          { fmt.Println("///////////////////////////"); $$ = []core.RextMetricsExtractor{ $1 } }
        | ssec EOL extblk
          { fmt.Println("jjjjjjjjjjj"); $$ = append($1, $3) }

srcsecsufixo : /* empty */
               { $$ = nil }
             |
               srcsecsufix
               { 
                 fmt.Println("srcsecsufixsrcsecsufixsrcsecsufix")
                 $$ = $1
               }
there is an error here
srcsecsufix : BLK optblkr sseco EOB
             {
              //  fmt.Println("aaaaaaaaaaaaaa", $3)
              //   // FIXME: Error handling
              //   // _ = util.MergeStoresInplace(dsGetOptions(), $2)
              //   $$ = $3
             }

srcsec  : srcverb IDENTIFIER FROM RESOURCE_PATH srcsecsufixo
          {
            fmt.Println("ttttttttttttttttt")
            // fmt.Println("1111111111111")
            // env := getRootEnv().env
            // // TODO error handling
            // ds, _ := env.NewMetricsDatasource($2)
            // ds.SetMethod($1)
            // ds.SetResourceLocation($4)
            // $$ = newMainSrc(ds)
          }
defsec  : defverb IDENTIFIER AS STR_LITERAL optblko
          {
            fmt.Println("hhhhhhhhhhhhhhhhhhhh", $5)
            // env := getRootEnv()
            // if defverb == 'AUTH' {
            //   $$ = newMainDef($4, env.NewAuthStrategy($2, $5))
            // }
            // TODO: Error handling
            $$ = nil
            fmt.Println("end hhhhhhhhhhhhhhhhhhhhhh")
          }
optblko : /* empty */
          { fmt.Println("LLRRRRRRRRRLLLLLLLL"); $$ = nil }
        | BLK optsblk EOB
          { fmt.Println("LLLLLLLLLL"); $$ = $2 }
// stkcls  : 'FOR STACK' idlst
//           { $$ = $2 }
// stkclso : /* empty */
//           { $$ = nil }
//         | stkcls EOL
//           { $$ = $1 }


srvstk  : FOR_STACK idlst
          { $$ = $2 }

srvstko : /* empty */
          { $$ = nil }
        | srvstk EOL
          { $$ = $1 }

srvcls  : FOR_SERVICE idlst
          { fmt.Println("1111110000010101000000000000000000000000000000000011111111111111111"); $$ = $2 }
srvclso : /* empty */
          { $$ = nil }
        | srvcls EOL
          { $$ = $1 }
mainsec : defsec
          { fmt.Println("mainseeeeeeeecc 1"); $$ = $1 }
        | srcsec
          { fmt.Println("mainseeeeeeeecc 2"); $$ = $1 }
mainblk : /* empty */
          { fmt.Println("mainnnnnnnnnnnnnnblk 1"); $$ = nil /*[]mainSecTuple { $1 }*/ }
        | mainblk mainsec
          { fmt.Println("mainnnnnnnnnnnnnnblk 2"); $$ = nil; /*append($1, $3)*/ }
eolo    : /* empty */
        | EOL
dataset : /*CTX*/ eolo DATASET BLK srvclso srvstko optblkr mainblk //EOB eolo
          {
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
