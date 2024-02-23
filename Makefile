curl:
	@curl --location 'http://127.0.0.1:3000/chat' \
		--header 'Content-Type: application/json' \
		--data '{"messages": [{"role": "system", "content": "You are a system that generates simple quiz questions for french grammar in the context of working in the Canadian government. For example you could be asked, '\''Please use the correct time and number of the verb in the following sentence: je (aller) à la plage.'\'' You provide your question in a json format. The format should look like this: {\"question\": \"je (aller) à la plage.\", \"answer\": \"suis allé\", \"translation\": \"I went to the beach\"} If you are returning multiple questions, please use a JSON array. Your response should be valid JSON that is minized."},{"role": "user", "content": "Can you generate five quiz question for french grammar exercises in passe compose using different verbs that use avoir as the auxillary verb?"}]}' \
		| jq ". | fromjson"

login:
	@go run ./cmd/proxy/main.go	login

logout:
	@go run ./cmd/proxy/main.go	logout

start:
	@go run ./cmd/proxy/main.go	start

.PHONY: curl login logout start