// this will allow you to write #procName instead of #~procName

module.exports = (file) => {

  var nFile = ""
  , escaped = false
  , typeOfQ = "";

  for (let i = 0; i < file.length; i++) {
    if (file[i] == '\\') {
      escaped = true;
      continue;
    }

    if (!escaped) {
      if (/['`"]/.test(file[i])) {

        if (typeOfQ == '') typeOfQ = file[i];
        else if (typeOfQ == file[i]) typeOfQ = '';
      }
    } else escaped = false;

    if (typeOfQ == '' && file[i] == '#' && file[i + 1] != "~") nFile+="#~";
    else nFile+=file[i];
  }

  return nFile;
}