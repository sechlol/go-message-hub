echo "Getting dependencies..."
go get "github.com/oleiade/lane"
go get "gopkg.in/fatih/set.v0"
go get "github.com/stretchr/testify/assert"

echo "Building executables..."
go build ./runnables/start_server.go
go build ./runnables/start_simulation.go
echo "Done!"