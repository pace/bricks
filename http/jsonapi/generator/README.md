# Limitations of the Api Generation

- Security Schemes of type _apiKey_ should not use the _Authorization_-Header, if more than one security scheme is used for any endpoint. 
Otherwise it is not possible to choose the right Authorization scheme for each request
