# Heavenly Dragons Authentication Server

## Setting up the project
### Environment
Create a `.env` file from the `.env.template`
### Database migration
run
```sh
go run github.com/prisma/prisma-client-go migrate dev
```
### Generating the prisma client
run
```
go run github.com/prisma/prisma-client-go generate
```
### Installing the dependencies
run
```
go get
```

### Running the code
Simply run 
```
go run ./main.go
```

