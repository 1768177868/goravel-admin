package services

import (
"os"
"testing"

"github.com/goravel/framework/facades"

_ "goravel/tests"
)

func TestMain(m *testing.M) {
if err := facades.Artisan().Call("migrate"); err != nil {
panic(err)
}

exit := m.Run()

os.Exit(exit)
}
