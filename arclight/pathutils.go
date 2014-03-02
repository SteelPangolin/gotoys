package arclight

import "sort"

// Given a list of paths, return a list of paths that must
// be created to form a complete directory tree.
// Assumes the input list has been Clean()ed already,
// and no paths start with /.
func ImplicitDirs(paths []string) []string {
    sort.Strings(paths)
    var out []string
    prev := ""
    for _, path := range paths {
        for end := 0 ; end < len(path) ; end++ {
            if path[end] != '/' {
                continue
            }
            ancestor := path[:end]
            if ancestor <= prev {
                continue
            }
            out = append(out, ancestor)
        }
        prev = path
    }
    return out
}
