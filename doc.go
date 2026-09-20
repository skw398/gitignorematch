// Package gitignorematch matches paths against Git .gitignore patterns.
//
// It deliberately does not walk directories, discover .gitignore files, or
// inspect a Git repository. The caller supplies the rules and tells Match
// whether the path being tested is a directory.
package gitignorematch
