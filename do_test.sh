go test -cover ./message/
go test -cover ./mexsocket/
go test -cover ./statbucket/
go test -cover ./client/
go test -cover ./hub/idpool/
go test -cover ./hub/
go test

go test -bench=. ./message/
go test -bench=. ./mexsocket/
go test -bench=. ./hub/idpool/
