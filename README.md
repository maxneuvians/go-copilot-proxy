# Go proxy for GitHub Copilot
The purpose of this application is to act as a proxy server for GitHub Copilot. It requires the user to have a valid GitHub copilot subscription. This is for research purposes only.


## Usage without binary 

```bash
make login
```

This will create a local file containing the GitHub Copilot token in the root directory of the project after you have logged in using your GitHub account and a device authentication code.

```bash
make start
```

This will start the server on port 3000. You can now use the server as a proxy for GitHub Copilot.

```bash
@curl --location 'http://127.0.0.1:3000/chat' \
		--header 'Content-Type: application/json' \
		--data '{"messages": [{"role": "system", "content": "You are a system that generates simple quiz questions for french grammar in the context of working in the Canadian government. For example you could be asked, '\''Please use the correct time and number of the verb in the following sentence: je (aller) à la plage.'\'' You provide your question in a json format. The format should look like this: {\"question\": \"je (aller) à la plage.\", \"answer\": \"suis allé\", \"translation\": \"I went to the beach\"} If you are returning multiple questions, please use a JSON array. Your response should be valid JSON that is minized."},{"role": "user", "content": "Can you generate five quiz question for french grammar exercises in passe compose using different verbs that use avoir as the auxillary verb?"}]}' \
		| jq ". | fromjson"
```

Will result in:

```
[
  {
    "question": "Hier, nous (acheter) un nouveau drapeau pour le bâtiment du gouvernement.",
    "answer": "avons acheté",
    "translation": "Yesterday, we bought a new flag for the government building."
  },
  {
    "question": "Le Premier ministre (visiter) plusieurs provinces l'année dernière.",
    "answer": "a visité",
    "translation": "The Prime Minister visited several provinces last year."
  },
  {
    "question": "Ils (répondre) à toutes les questions lors de la conférence de presse.",
    "answer": "ont répondu",
    "translation": "They answered all the questions at the press conference."
  },
  {
    "question": "J' (envoyer) le rapport à mon superviseur hier.",
    "answer": "ai envoyé",
    "translation": "I sent the report to my supervisor yesterday."
  },
  {
    "question": "Tu (signer) le document avant de quitter le bureau?",
    "answer": "as signé",
    "translation": "Did you sign the document before leaving the office?"
  }
]
```

```bash
make logout
```

This will remove the local file containing the GitHub Copilot token in the root directory of the project.