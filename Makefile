coverage:
	go test ./... -coverprofile=coverage.out

report:
	go tool cover -html=coverage.out
