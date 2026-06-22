![logo](.github/images/urban-dict.png)

# Urban dict

This API retrieves entries from Urban Dictionary, and return them formatted as plain text

## Getting Started

### StreamElements
Add a custom command using the `${customapi.url}`, like the example bellow

```sh
${customapi.https://darckfast.com/api/urban?term=${pathescape ${0:}}&channel=$(channel)}
```

### NightBot
Add a custom command using the `$(urlfetch url)`, like the example bellow
```sh
$(urlfetch https://darckfast.com/api/urban?term=$(querystring)&channel=$(channel))
```

### FossaBot
Add a custom command using the `$(customapi url)`, like the example bellow
```sh
$(customapi https://darckfast.com/api/urban?term=$(querystring)&channel=$(channel))
```

### MooBot
Create a `URL-Fetch` command, like the example bellow
```sh
https://darckfast.com/api/urban?term=Command arguments&channel=Username of the channel
```

### cURL
The API can be called directly with a GET HTTP request

```sh
curl https://darckfast.com/api/urban \\
    --url-query 'term=my term' \\
    --url-query 'channel=my channel name'
```

[Check the complete documentation for more](https://darckfast.com/docs/urban)


