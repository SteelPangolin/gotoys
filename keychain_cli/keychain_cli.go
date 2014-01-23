package main

import "fmt"
import "github.com/SteelPangolin/gotoys/keychain"

func main() {
    version, err := keychain.GetVersion()
    if err != nil {
        fmt.Printf("Keychain Services error %#v\n", err)
    } else {
        fmt.Printf("Keychain Services version %#v\n", version)
    }
    
    kc, err := keychain.Open("imaginarykeychain")
    if err != nil {
        fmt.Printf("Keychain Services error %#v\n", err)
    } else {
        fmt.Printf("Opened keychain %#v\n", kc)
    }
}
