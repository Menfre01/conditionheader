# condition-header

condition-header is a plugin for [traefik](https://traefik.io) that allows you to conditionally add headers to responses.

## Configuration

Add plugin:
```yaml
experimental:
  plugins:
    condition-header:
      moduleName: github.com/Menfre01/conditionheader
      version: v0.0.2
```

Configure middleware:
```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: add-response-header
spec:
  plugin:
    condition-header:
      rules:
        - conditions:
            Content-Type: text/html.*
          headers:
            Cache-Control: no-cache, must-revalidate
```

