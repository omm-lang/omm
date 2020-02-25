package main

import "strings"
import "fmt"
import "encoding/json"

// #cgo CFLAGS: -std=c99
import "C"

//export Multiply
func Multiply(_num1P *C.char, _num2P *C.char, calc_paramsP *C.char, line_ C.int) *C.char {

  _num1 := C.GoString(_num1P)
  _num2 := C.GoString(_num2P)
  calc_params_str := C.GoString(calc_paramsP)

  line := int(line_)

  _ = line

  var calc_params paramCalcOpts

  _ = json.Unmarshal([]byte(calc_params_str), &calc_params)

  if returnInit(_num1) == "0" || returnInit(_num2) == "0" {
    return C.CString("0")
  }

  decIndex := 0

  if strings.Contains(_num1, ".") {
    decIndex+=len(_num1) - strings.Index(_num1, ".")
  }
  if strings.Contains(_num2, ".") {
    decIndex+=len(_num1) - strings.Index(_num2, ".")
  }

  decIndex--

  _num1 = strings.Replace(_num1, ".", "", 1)
  _num2 = strings.Replace(_num2, ".", "", 1)

  if isLess(_num1, _num2) {
    _num1, _num2 = _num2, _num1
  }

  neg := false

  if strings.HasPrefix(_num1, "-") {
    _num1 = _num1[1:]
    neg = !neg
  }
  if strings.HasPrefix(_num2, "-") {
    _num2= _num2[1:]
    neg = !neg
  }

  var nNum string

  if isLess(_num2, calc_params.mult_thresh) {
    nNum = "0"

    for ;returnInit(_num2) != "0"; {
      nNum = C.GoString(Add(C.CString(nNum), C.CString(_num1), calc_paramsP, line_))
      _num2 = C.GoString(Subtract(C.CString(_num2), C.CString("1"), calc_paramsP, line_))

      if calc_params.logger {
        fmt.Println("Omm Logger ~ Multiplication: " + nNum)
      }
    }

    if decIndex > -1 {

      nNum = Reverse(nNum)

      nNum = nNum[:decIndex] + "." + nNum[decIndex:]

      nNum = Reverse(nNum)
    }

    if neg == true {
      nNum = "-" + nNum
    }
  } else {
    nNum = "0"

    for i, o := len(_num2) - 1, 0; i >= 0; i, o = i - 1, o + 1 {
      nNum = C.GoString(Add(C.CString(nNum), C.CString(C.GoString(Multiply(C.CString(string([]rune(_num2)[i])), C.CString(_num1), calc_paramsP, line_)) + RepeatAdd("0", o)), calc_paramsP, line_))

      if calc_params.logger {
        fmt.Println("Omm Logger ~ Multiplication: " + nNum)
      }
    }
  }

  return C.CString(returnInit(nNum))
}
