package compiler

import "io/ioutil"
import "path"
import "strings"
import "github.com/omm-lang/omm/oat/encoding"

import . "github.com/omm-lang/omm/lang/types"

func includeSingle(filename string, line uint64, dir string, fromstd bool) ([]Action, CompileErr) {

  if fromstd {
    filename = path.Join(path.Join(Ommbasedir, "stdlib"), filename)
  }

  if strings.HasSuffix(filename, ".oat") {
    decoded, e := oatenc.OatDecode(filename, 1)

    if e != nil {
      return nil, _CompileErr{
        Msg: e.Error(),
        FName: filename,
        Line: line,
      }
    }

    return decoded["$main"], nil
  }

  if strings.HasSuffix(filename, ".omm") {
    filename = strings.TrimSuffix(filename, ".omm")
  }

  filename+=".omm"

  for _, v := range included {
    if v == filename {
      return []Action{}, nil
    }
  }

  content, err := ioutil.ReadFile(filename)

  included = append(included, filename)

  if err != nil {
    return []Action{}, makeCompilerErr("Could not find file: " + filename, dir, line)
  }

  compiled, e := inclCompile(string(content), filename, false)

  if e != nil {
    return []Action{}, e
  }

  return compiled, nil
}

func includer(filename string, line uint64, dir string, fromstd bool) ([][]Action, CompileErr) {

  if strings.HasSuffix(filename, "*") {

    if fromstd {
      filename = path.Join(path.Join(Ommbasedir, "stdlib"), filename)
    }

    files, e := ioutil.ReadDir(strings.TrimSuffix(filename, "*"))

    if e != nil {
      return [][]Action{}, makeCompilerErr("Could not find directory: " + filename, dir, line)
    }

    var actions [][]Action

    for _, v := range files {

      if !strings.HasSuffix(v.Name(), ".omm") && !strings.HasSuffix(v.Name(), ".oat") { //if it is not an omm or an oat file, skip it
        continue
      }

      if v.IsDir() {
        inc, e := includer(path.Join(strings.TrimSuffix(filename, "*"), v.Name() + "/*"), line, dir, false)

        if e != nil {
          return [][]Action{}, e
        }

        actions = append(actions, inc...)
      } else {
        inc, e := includeSingle(path.Join(strings.TrimSuffix(filename, "*"), v.Name()), line, dir, false)

        if e != nil {
          return [][]Action{}, e
        }

        actions = append(actions, inc)
      }
    }

    return actions, nil
  } else {
    inc, e := includeSingle(filename, line, dir, fromstd)

    if e != nil {
      return [][]Action{}, e
    }

    return [][]Action{ inc }, nil
  }
}