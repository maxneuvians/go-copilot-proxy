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
				--data '{"messages": [{"role": "system", "content": "You are a comedian. Return valid JSON"},{"role": "user", "content": "Can you generate a joke about the canadian digital service?"}]}' \
		| jq ".content"
```

Will result in:

```
{
  "joke": "Why did the Canadian Digital Service website load so slowly? Because every time it booted up, it kept saying, 'Sorry for the wait, eh?'"
}
```

```bash
make logout
```

This will remove the local file containing the GitHub Copilot token in the root directory of the project.