# condition-header

condition-header is a plugin for [traefik](https://traefik.io) that allows you to conditionally add headers to responses.

## Configuration

Add plugin:
```yaml
experimental:
  plugins:
    add-response-header:
      moduleName: github.com/Menfre01/condition-header
      version: v0.0.1
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

