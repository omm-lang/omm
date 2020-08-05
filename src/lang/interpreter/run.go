package interpreter

import "os"
import "fmt"
import "strings"

import . "lang/types"

type overload_after struct {
  name   string
  val  []Action
}

var threads []OmmThread

//export RunInterpreter
func RunInterpreter(compiledVars map[string][]Action, cli_params CliParams) {

  var dirnameOmmStr OmmString
  dirnameOmmStr.FromGoType(cli_params.Directory)
  var dirnameOmmType OmmType = dirnameOmmStr

  var instance Instance

  instance.Params = cli_params
  instance.vars = make(map[string]*OmmVar)
  instance.globals = make(map[string]*OmmVar)

  instance.vars["$__dirname"] = &OmmVar{
    Name: "$__dirname",
    Value: &dirnameOmmType,
  }

  var doafter = make([]overload_after, 0)

  for k, v := range compiledVars {

    if strings.HasPrefix(k, "ovld/") {
      //Using this, because the order of the map is not maintained, so this can cause a nil pointer
      doafter = append(doafter, overload_after{
        name: strings.TrimSpace(strings.TrimPrefix(k, "ovld/")),
        val: v,
      })
      continue
    }

    var global = OmmVar{
      Name: k,
      Value: instance.interpreter(v, []string{"at the global interpreter"}).Exp,
    }
    instance.globals[k] = &global
  }

  for _, v := range doafter {

    var _fn = *instance.globals[strings.TrimPrefix(v.name, "ovld/")].Value

    //if it not a function, force it to be one
    switch _fn.(type) {
      case OmmFunc: //ignore
      default:
        _fn = OmmFunc{
          Overloads: []Overload{},
        }
    }

    var fn = _fn.(OmmFunc)
    fn.Overloads = append(fn.Overloads, v.val[0].Value.(OmmFunc).Overloads[0])
    *instance.globals[v.name].Value = fn
  }

  instance.vars = instance.globals //copy the globals into the vars

  for k, v := range GoFuncs {
    var gofunc OmmType = OmmGoFunc{
      Function: v,
    }
    instance.vars["$" + k] = &OmmVar{
      Name: "$" + k,
      Value: &gofunc,
    }
  }

  if _, exists := instance.vars["$main"]; !exists {
    fmt.Println("Given program has no entry point/main function")
    os.Exit(1)
  } else {

    switch (*instance.vars["$main"].Value).(type) {
      case OmmFunc:
        main := instance.vars["$main"]

        calledP := instance.interpreter((*main.Value).(OmmFunc).Overloads[0].Body, []string{"at the entry caller"}).Exp

        if calledP == nil {
          os.Exit(0)
        }

        called := *calledP

        for _, v := range threads {
          v.WaitFor()
        }

        var exitType int64 = 0

        switch called.(type) {
          case OmmNumber:
            exitType = int64(called.(OmmNumber).ToGoType())
        }

        os.Exit(int(exitType))
      default:
        fmt.Println("Entry point was not given as a function")
        os.Exit(1)
    }
  }
}
