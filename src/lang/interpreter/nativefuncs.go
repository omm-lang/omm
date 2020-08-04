package interpreter

//all of the gofuncs
//functions written in go that are used by omm

import "bufio"
import "os"
import "fmt"
import "time"
import "strconv"
import "strings"
import "os/exec"
import "runtime"

import . "lang/types"

type OmmGoFunc struct {
  Function func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType
}

func (ogf OmmGoFunc) Format() string {
  return "{native gofunc}"
}

func (ogf OmmGoFunc) Type() string {
  return "gofunc"
}

func (ogf OmmGoFunc) TypeOf() string {
  return ogf.Type()
}


//export GoFuncs
var GoFuncs = map[string]func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {
  "input": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    scanner := bufio.NewScanner(os.Stdin)

    if len(args) == 0 {
      //if it has 0 or 1 arg, there is no error
    } else if len(args) == 1 {

      switch (*args[0]).(type) {
        case OmmString:
          str := (*args[0]).(OmmString).ToGoType()
          fmt.Print(str + ": ")
        default:
          OmmPanic("Expected a string as the argument to input[]", line, file, stacktrace)
      }

    } else {
      OmmPanic("Function input requires a parameter count of 0 or 1", line, file, stacktrace)
    }

    //get user input and convert it to OmmType
    scanner.Scan()
    input := scanner.Text()
    var inputOmmType OmmString
    inputOmmType.FromGoType(input)
    var inputType OmmType = inputOmmType

    return &inputType
  },
  "typeof": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) != 1 {
      OmmPanic("Function typeof requires a parameter count of 1", line, file, stacktrace)
    }

    typeof := (*args[0]).TypeOf()

    var str OmmString
    str.FromGoType(typeof)

    //convert to OmmType interface
    var ommtype OmmType = str

    return &ommtype
  },
  "defop": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) != 4 {
      OmmPanic("Function defop requires a parameter count of 4", line, file, stacktrace)
    }

    if (*args[0]).Type() != "string" || (*args[1]).Type() != "string" || (*args[2]).Type() != "string" || (*args[3]).Type() != "function" {
      OmmPanic("Function defop requires [string, string, string, function]", line, file, stacktrace)
    }

    operation := (*args[0]).(OmmString).ToGoType()
    operand1 := (*args[1]).(OmmString).ToGoType()
    operand2 := (*args[2]).(OmmString).ToGoType()
    function := (*args[3]).(OmmFunc)

    if len(function.Overloads[0].Params) != 2 {
      OmmPanic("Expected a parameter count of 2 for the fourth argument of defop", line, file, stacktrace)
    }

    Operations[operand1 + " " + operation + " " + operand2] = func(val1, val2 OmmType, instance *Instance, stacktrace []string, line uint64, file string) *OmmType {
      *instance.vars[function.Overloads[0].Params[0]] = OmmVar{
        Name: function.Overloads[0].Params[0],
        Value: &val1,
      }
      *instance.vars[function.Overloads[0].Params[1]] = OmmVar{
        Name: function.Overloads[0].Params[1],
        Value: &val2,
      }

      return instance.interpreter(function.Overloads[0].Body, stacktrace).Exp
    }

    var tmpundef OmmType = undef
    return &tmpundef //return undefined
  },
  "append": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) != 2 {
      OmmPanic("Function append requires a parameter count of 2", line, file, stacktrace)
    }

    if (*args[0]).Type() != "array" {
      OmmPanic("Function append requires the first argument to be an array", line, file, stacktrace)
    }

    appended := append((*args[0]).(OmmArray).Array, args[1])
    var arr OmmType = OmmArray{
      Array: appended,
      Length: uint64(len(appended)),
    }

    return &arr
  },
  "prepend": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) != 2 {
      OmmPanic("Function prepend requires a parameter count of 2", line, file, stacktrace)
    }

    if (*args[0]).Type() != "array" {
      OmmPanic("Function prepend requires the first argument to be an array", line, file, stacktrace)
    }

    prepended := append([]*OmmType{ args[1] }, (*args[0]).(OmmArray).Array...)
    var arr OmmType = OmmArray{
      Array: prepended,
      Length: uint64(len(prepended)),
    }

    return &arr
  },
  "exit": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) == 1 {

      switch (*args[0]).(type) {
        case OmmNumber:

          var gonum = (*args[0]).(OmmNumber).ToGoType()
          os.Exit(int(gonum))

        case OmmBool:

          if (*args[0]).(OmmBool).ToGoType() == true {
            os.Exit(0)
          } else {
            os.Exit(1)
          }

        default:
          os.Exit(0)
      }

    } else if len(args) == 0 {
      os.Exit(0)
    } else {
      OmmPanic("Function exit requires a parameter count of 1 or 0", line, file, stacktrace)
    }

    var tmpundef OmmType = undef
    return &tmpundef
  },
  "wait": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) == 1 {

      if (*args[0]).Type() != "number" {
        OmmPanic("Function wait requires a number as the argument", line, file, stacktrace)
      }

      var amt = (*args[0]).(OmmNumber)

      var n4294967295 = zero
      n4294967295.Integer = &[]int64{ 5, 9, 2, 7, 6, 9, 4, 9, 2, 4 }

      //if amt is less than 2 ^ 32 - 1, just convert to a go int
      if isLess(amt, n4294967295) {
        gonum := amt.ToGoType()

        time.Sleep(time.Duration(gonum) * time.Millisecond)
      } else {
        //this is how this works
        /*
          the loop starts at 0 with an increment of 2 ^ 32 - 1
          in each iteration, it will wait for 2 ^ 32 - 1 milliseconds
        */

        for i := zero; isLess(i, amt); i = (*number__plus__number(i, n4294967295, &Instance{}, stacktrace, line, file)).(OmmNumber) {
          time.Sleep(4294967295 * time.Millisecond)
        }
      }

    } else {
      OmmPanic("Function wait requires a parameter count of 1", line, file, stacktrace)
    }

    var tmpundef OmmType = undef
    return &tmpundef
  },
  "make": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) == 1 {

      switch (*args[0]).(type) {
      case OmmProto:
          var ommtype OmmType = OmmObject{}.New((*args[0]).(OmmProto))
          return &ommtype
        default:
          OmmPanic("Function make requires a structure as the argument", line, file, stacktrace)
      }

    } else {
      OmmPanic("Function make requires a parameter count of 1", line, file, stacktrace)
    }

    var tmpundef OmmType = undef
    return &tmpundef
  },
  "len": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) != 1 {
      OmmPanic("Function len requires a parameter count of 1", line, file, stacktrace)
    }

    switch (*args[0]).(type) {
      case OmmString:
        var length = (*args[0]).(OmmString).Length

        var ommnumber_length OmmNumber
        ommnumber_length.FromString(strconv.FormatUint(length, 10))
        var ommtype OmmType = ommnumber_length
        return &ommtype

      case OmmArray:
        var length = (*args[0]).(OmmArray).Length

        var ommnumber_length OmmNumber
        ommnumber_length.FromString(strconv.FormatUint(length, 10))
        var ommtype OmmType = ommnumber_length
        return &ommtype

      case OmmHash:
        var length = (*args[0]).(OmmHash).Length

        var ommnumber_length OmmNumber
        ommnumber_length.FromString(strconv.FormatUint(length, 10))
        var ommtype OmmType = ommnumber_length
        return &ommtype

    }

    var tmpzero OmmType = zero
    return &tmpzero
  },
  "clone": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) != 1 {
      OmmPanic("Function len requires a parameter count of 1", line, file, stacktrace)
    }

    val := *args[0]

    switch val.(type) {

      case OmmArray:

        var arr = val.(OmmArray).Array
        var cloned = append([]*OmmType{}, arr...) //append it to nothing (to clone it)
        var ommtype OmmType = OmmArray{
          Array: cloned,
          Length: val.(OmmArray).Length,
        }

        return &ommtype

      case OmmBool:

        //take inderect of the bool, and place it in a temporary variable
        var tmp = *val.(OmmBool).Boolean

        var returner OmmType = OmmBool{
          Boolean: &tmp, //take address of tmp and place it into `Boolean` field of returner
        }

        return &returner

      case OmmHash:
        var hash = val.(OmmHash).Hash

        //clone it into `cloned`
        var cloned = make(map[string]*OmmType)
        for k, v := range hash {
          cloned[k] = v
        }
        ////////////////////////

        var ommtype OmmType = OmmHash{
          Hash: cloned,
          Length: val.(OmmHash).Length,
        }

        return &ommtype

      case OmmNumber:
        var number = val.(OmmNumber)

        //copy the integer and decimal
        var integer = append([]int64{}, *number.Integer...)

        var decimal []int64

        if number.Decimal != nil {
          decimal = append([]int64{}, *number.Decimal...)
        }
        //////////////////////////////

        var newnum OmmType = OmmNumber{
          Integer: &integer,
          Decimal: &decimal,
        }
        return &newnum

      case OmmObject:

        //almost the same as hash (just clone the instance)
        var instance = val.(OmmObject).Instance

        //clone it into `cloned`
        var cloned = make(map[string]*OmmType)
        for k, v := range instance {
          cloned[k] = v
        }
        ////////////////////////

        var ommtype OmmType = OmmObject{
          Name: val.(OmmObject).Name,
          Instance: cloned,
        }

        return &ommtype

      case OmmRune:

        //take inderect of the rune, and place it in a temporary variable
        var tmp = *val.(OmmRune).Rune

        var returner OmmType = OmmRune{
          Rune: &tmp, //take address of tmp and place it into `Rune` field of returner
        }

        return &returner

      case OmmString:

        var tmp = val.(OmmString).ToGoType() //convert it to a go type

        var returner OmmType = OmmString{
          String: &tmp,
          Length: val.(OmmString).Length,
        }

        return &returner

      default:
        OmmPanic("Cannot clone type \"" + val.Type() + "\"", line, file, stacktrace)
    }

    var tmpundef OmmType = undef
    return &tmpundef
  },
  "panic": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) != 1 {
      OmmPanic("Function panic requires a parameter count of 1", line, file, stacktrace)
    }
    if (*args[0]).Type() != "string" {
      OmmPanic("Function panic requires the argument to be a string", line, file, stacktrace)
    }

    var err = (*args[0]).(OmmString)

    OmmPanic(err.ToGoType(), line, file, stacktrace)

    var tmpundef OmmType = undef
    return &tmpundef
  },
  "exec": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) == 1 {
      if (*args[0]).Type() != "string" {
        OmmPanic("Function exec requires the argument to be a string", line, file, stacktrace)
      }

      var cmd = (*args[0]).(OmmString).ToGoType()

      command := exec.Command(cmd)
      out, _ := command.CombinedOutput()

      var stringValue OmmString
      stringValue.FromGoType(string(out))
      var ommtype OmmType = stringValue
      return &ommtype
    } else if len(args) == 2 {
      if (*args[0]).Type() != "string" || (*args[1]).Type() != "string" {
        OmmPanic("Function exec requires both arguments to be strings", line, file, stacktrace)
      }

      var cmd = (*args[0]).(OmmString).ToGoType()
      var stdin = (*args[1]).(OmmString).ToGoType()

      command := exec.Command(cmd)
      command.Stdin = strings.NewReader(stdin)
      out, _ := command.CombinedOutput()

      var stringValue OmmString
      stringValue.FromGoType(string(out))
      var ommtype OmmType = stringValue
      return &ommtype
    } else {
      OmmPanic("Function exec requires a parameter count of 1 or 2", line, file, stacktrace)
    }

    var tmpundef OmmType = undef
    return &tmpundef
  },
  "chdir": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {

    if len(args) != 1 {
      OmmPanic("Function chdir requires a parameter count of 1", line, file, stacktrace)
    }
    if (*args[0]).Type() != "string" {
      OmmPanic("Function chdir requires the argument to be a string", line, file, stacktrace)
    }

    var dir = (*args[0]).(OmmString).ToGoType()

    os.Chdir(dir)

    instance.vars["$__dirname"] = &OmmVar{ //change the __dirname variable too
      Name: "$__dirname",
      Value: args[0],
    }

    var tmpundef OmmType = undef
    return &tmpundef
  },
  "getos": func(args []*OmmType, stacktrace []string, line uint64, file string, instance *Instance) *OmmType {
    var os = runtime.GOOS
    var ommstr OmmString
    ommstr.FromGoType(os);
    var ommtype OmmType = ommstr
    return &ommtype
  },
  //osm functions
  "osm.start": OSM_start,
  "osm.handle": OSM_handle,
  ///////////////
}