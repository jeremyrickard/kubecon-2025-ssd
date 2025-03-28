package main

package main

import (
        _ "embed"
        "os"

        colorable "github.com/mattn/go-colorable"
        log "github.com/sirupsen/logrus"
        "github.com/jeremyrickard/cmd"
)

//go:embed constraints.yml
var constraintsYml string

func main() {
        log.SetFormatter(&log.TextFormatter{ForceColors: true})
        log.SetOutput(colorable.NewColorableStdout())
        if err := cmd.NewRootCmd().Execute(); err != nil {
                os.Exit(1)
        }
        cmd.ConstraintsFile = constraintsYml
}