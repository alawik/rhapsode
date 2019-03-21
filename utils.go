package main

import (
    "os"
    "path/filepath"
)

func DoesPathExist(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return true, err
}

func LangOfFunc(fileName string) (lang string) {
    //extension := filepath.Ext(fileName)
    extLangMap := map[string]string{
        ".js":   "node",
        ".go":   "go",
        ".rb":   "ruby",
        ".py":   "python",
        ".php":  "php",
        ".jl":   "julia",
        ".java": "java",
    }
    return extLangMap[filepath.Ext(fileName)]

    /* When lang of file is not found in map
    if val, ok := extLangMap[filepath.Ext(fileName)]; ok {
        //do something here
    }
    */
}
