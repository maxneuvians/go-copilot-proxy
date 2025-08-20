curl:
	@curl --location 'http://127.0.0.1:3000/chat' \
		--header 'Content-Type: application/json' \
		--data '{"model": "o3-mini", "messages": [{"role": "system", "content": "You are a comedian. Return valid JSON"},{"role": "user", "content": "Can you generate a joke about the canadian digital service?"}]}' \
		| jq .

login:
	@go run ./cmd/proxy/main.go	login

logout:
	@go run ./cmd/proxy/main.go	logout

start:
	@go run ./cmd/proxy/main.go	start

.PHONY: curl login logout start