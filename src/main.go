package main

// go-grep
// basic grep implementation in Go

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
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

// remaining CLI flags
var argCount    *bool = flag.Bool("c", false, "write only count of lines to std-out")
var argNames    *bool = flag.Bool("l", false, "write only the names of the files containing selected lines")
var argQuiet    *bool = flag.Bool("q", false, "write nothing to std-out, regardless of matching lines")
var argCase     *bool = flag.Bool("i", false, "perform pattern matching without regard to case")
var argLines    *bool = flag.Bool("n", false, "precede each output line by its line number in the file")
var argSuppress *bool = flag.Bool("s", false, "suppress error messages for nonexistent or unreadable files") 
var argNoMatch  *bool = flag.Bool("v", false, "select only lines that do not match the pattern(s)")
var argEntire   *bool = flag.Bool("x", false, "select only lines that use all characters to match the pattern")

func main () { 

  flag.Var(&argPatternList, "e", "read one or more patterns from std-in")
  flag.Var(&argPatternFiles, "f", "read one or more patterns from file")
  flag.Parse()

  if flag.NArg() == 0 && argPatternList == nil && argPatternFiles == nil { 
    fmt.Println("Usage: go-grep [OPTIONS...] PATTERNS [FILE...]")
    os.Exit(0)
  }

  // get pattern(s) supplied with -e
  var inputPatterns []string = argPatternList

  // get pattern(s) supplied with -f
  for _,fname := range argPatternFiles { 
    file := openFile(fname); defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan() { 
      inputPatterns = append(inputPatterns, scanner.Text())
    }
    if err := scanner.Err(); err != nil && !*argSuppress { log.Fatal(err) }
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
    if grepFile(openFile(fname), inputPatterns) && *argNames { 
      fmt.Println(fname)
    }
  }

  // if no files, read from std-in
  if len(inputFiles) == 0 {
    if grepFile(os.Stdin, inputPatterns) && *argNames { 
      fmt.Println("(standard input)")
    }
  }

}

func openFile(fname string) *os.File {
  // opens filename with error handling
  file, err := os.Open(fname)
  if err != nil && !*argSuppress { log.Fatal(err) }
  return file
}

func matchLine(text string, patterns []string) bool { 
  /* checks if text contains a match for any pattern within patterns
  returns true if so, otherwise false
  */
  var isMatch bool = false
  for _,pattern := range patterns { 

    // is pattern case-insensitive?
    if *argCase { pattern = "(?i)" + pattern }

    regex, err := regexp.Compile(pattern)
    if err != nil { log.Fatal(err) }
    matched := regex.MatchString(text)

    // does pattern occupy entire line? 
    if *argEntire && regex.FindString(pattern) != text { matched = false } 

    if matched { isMatch = true; break } 
  }
  if *argNoMatch { isMatch = !isMatch } 
  return isMatch
}

func grepFile(file *os.File, patterns []string) bool { 
  /* checks if each line in file matches any pattern in patterns
  if true, prints the line according to CLI arguments
  if at least one line in file matches, returns true, else false
  */
  var isMatch bool = false
  var isPrint bool = !*argQuiet && !*argNames && !*argCount
  var text string = "" 
  var lineNum, matchCount uint = 0, 0

  defer file.Close()
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {

    lineNum++
    text = scanner.Text()

    if matchLine(text, patterns) {
      isMatch = true; matchCount++
      if *argLines && isPrint { fmt.Printf("%v:", lineNum) }
      if isPrint { fmt.Println(text) }

    }
  }
  if *argCount { fmt.Println(matchCount) }
  if err := scanner.Err(); err != nil && !*argSuppress { log.Fatal(err) }
  return isMatch
}
