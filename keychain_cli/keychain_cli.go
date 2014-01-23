package main

import "fmt"
import "github.com/SteelPangolin/gotoys/keychain"

func main() {
    version, err := keychain.GetVersion()
    if err != nil {
        fmt.Printf("Keychain Services error %#v\n", err)
        return
    }
    fmt.Printf("Keychain Services version %#v\n", version)
}
