![logo](.github/images/urban-dict.png)

# Urban dict

This API retrieves entries from Urban Dictionary, and return them formatted as plain text

It also has a 10 seconds cache for random entries ( without the term ), meaning it's essential to pass the argument `channel` in the URL

The max characters are capped at 400

## StreamElements
To set up this using StreamElements, add a custom command with the following 

```
${customapi.https://darckfast.com/api/urban?term=${pathescape ${0:}}&channel=$(channel)}
```

### Usage

A term can be informed or not, in case it's empty, a random entry will be returned

```
!urban
!urban glizzy
```
