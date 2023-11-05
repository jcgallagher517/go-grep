package main

import (
	"fmt"
	"log"
	"os"
  "bufio"
	// "regexp"
	"flag"
)

// custom type to accommodate multiple occurrences of -e and -f arguments
type argList []string
func (*argList) String() string { return "" } 
func (args *argList) Set(value string) error { 
  *args = append(*args, value)
  return nil
} 
var argPatternList, argPatternFiles argList

// remaining CLI options
var argCount    *bool = flag.Bool("c", false, "write only count of lines to std-out")
var argNames    *bool = flag.Bool("l", false, "write only the names of the files containing selected lines")
var argQuiet    *bool = flag.Bool("q", false, "write nothing to std-out, regardless of matching lines")
var argCase     *bool = flag.Bool("i", false, "perform pattern matching without regard to case")
var argLines    *bool = flag.Bool("n", false, "precede each output line by its line number in the file")
var argSuppress *bool = flag.Bool("s", false, "suppress error messages for nonexistent or unreadable files") 
var argNoMatch  *bool = flag.Bool("v", false, "select only lines that do not match the pattern(s)")
var argEntire   *bool = flag.Bool("x", false, "select only lines that use all characters to match the pattern")

func matchLine(text string, patterns *[]string) bool { 
  /* Describe function here
  returns true if text matches any pattern, false otherwise
  */
  fmt.Println(text)
  return true
}

func main () { 

  flag.Var(&argPatternList, "e", "read one or more patterns from std-in")
  flag.Var(&argPatternFiles, "f", "read one or more patterns from file")
  flag.Parse()

  if flag.NArg() == 0 { 
    fmt.Println("Usage: go-grep [OPTIONS...] PATTERNS [FILE...]")
    os.Exit(0)
  }

  // get pattern(s) supplied with -e
  var inputPatterns []string = argPatternList

  // get pattern(s) supplied with -f
  addPattern := func(p string, ps *[]string) bool { 
    *ps = append(*ps, p)
    return true
  }
  for _,fname := range argPatternFiles { 
    actOnFile(openFile(fname), addPattern, &inputPatterns)
  }

  // get input filename(s) and/or lingering solo pattern
  var inputFiles []string
  if inputPatterns == nil { 
    inputPatterns = append(inputPatterns, flag.Arg(0))
    inputFiles = flag.Args()[1:]
  } else { 
    inputFiles = flag.Args()
  }

  // loop through files 
  for _,fname := range inputFiles {
    if actOnFile(openFile(fname), matchLine, &inputPatterns) && *argNames { 
      fmt.Println(fname)
    }
  }

  // if no files, read from std-in
  if inputFiles == nil {
    actOnFile(os.Stdin, matchLine, &inputPatterns)
    if *argNames { fmt.Println("(standard input)") } 
  }

}

func openFile(fname string) *os.File {
  // opens filename with error handling
  file, err := os.Open(fname)
  if err != nil && !*argSuppress { log.Fatal(err) }
  return file
}

func actOnFile(file *os.File, action func(text string, patterns *[]string) bool, patterns *[]string) bool { 
  /* applies action to each line of file 
  action accepts the line as its first input
  with patterns as remaining inputs
  and returns true if any lines match a pattern, false otherwise
  */
  defer file.Close()
  scanner := bufio.NewScanner(file)
  var isMatch bool = false
  for scanner.Scan() {
    if action(scanner.Text(), patterns) { isMatch = true }
  }
  if err := scanner.Err(); err != nil && !*argSuppress { log.Fatal(err) }
  return isMatch
}

