## Setup

```
go get github.com/harsh-98/witnetBOT
cd $GOPATH/src/github.com/harsh-98/witnetBOT
cp .witnet.yaml.env .witnet.yaml
```

Edit .witnet.yaml, adding user, password, db and host of mysql db. Load witnet.sql for creating the tables in the database.

## Starting the service

```
go run main.go --debug
```

or

```
go build main.go
./witnetBOT --debug
```
