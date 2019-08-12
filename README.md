Holdex Platform Backend Go Utility Library
=======================================

This public library contains methods that are useful for applications using Holdex Platform API.

## [Auth jwt](https://github.com/holdex/hp-backend-lib/src/master/auth/jwt.go)
##### Middleware functions to check authorization for HTTP request

## [Context](https://github.com/holdex/hp-backend-lib/src/master/ctx/ctx.go)
##### Functions to (set/get) specific information in context variables

## [Date util](https://github.com/holdex/hp-backend-lib/src/master/date/date.go) (check is two dates are equal )
`d1 := date.Date{Year: 2011, Month: 1, Day: 1}`<br/>
`d2 := date.Date{Year: 2011, Month: 1, Day: 1}`<br/>
`if libdate.Equal(d1, d2) {`<br/>
	&nbsp;&nbsp;&nbsp;&nbsp;`fmt.Print("dates are equal")`<br/>
`}`

## [Errors](https://github.com/holdex/hp-backend-lib/src/master/err)
##### Functions to create specific reason for errors

## [File validation](https://github.com/holdex/hp-backend-lib/src/master/file/file.go)
1. Parse multi-part form data files
2. Decode / Encode file

## [GRPC](https://github.com/holdex/hp-backend-lib/src/master/grpc)
##### Initializer function for grpc server
##### Oauth middleware

## [HTTP](https://github.com/holdex/hp-backend-lib/src/master/http/utils.go)
##### Functions to parse HTTP request/response, return custom errors

## [UUID](https://github.com/holdex/hp-backend-lib/src/master/id/id.go)
##### generate unique uuid, used for test practice
`id := libid.GenerateUniqueID()`

## [JWT](https://github.com/holdex/hp-backend-lib/src/master/jwt/service.go)
##### Service to create & parse jwt tokens

## [JSON](https://github.com/holdex/hp-backend-lib/src/master/json/json.go)
##### Functions to marshal/unmarshal json structure ignoring the error

## [Logger](https://github.com/holdex/hp-backend-lib/src/master/log/logger.go) 
##### Custom logger handler built on top of grpc logger

## [Password](https://github.com/holdex/hp-backend-lib/src/master/password/password.go)
##### Functions to encrypt password & check hash matching

## [Postgres](https://github.com/holdex/hp-backend-lib/src/master/pq/pq.go)
##### Initializer function for sql connection via postgres driver

## [Protobuf](https://github.com/holdex/hp-backend-lib/src/master/protobuf/protoc-gen-gogqlenum/main.go)
##### Binary plugin to generate GQL enum types from proto message 

## [Rollbar](https://github.com/holdex/hp-backend-lib/src/master/rollbar/rollbar.go)
#####  Function to setup rollbar 

## [Strings](https://github.com/holdex/hp-backend-lib/src/master/strings/strings.go)
##### Functions to check strings (length, equality, matching)

## [Sync](https://github.com/holdex/hp-backend-lib/src/master/sync/status.go)
##### Functions used for to check aggregate status
