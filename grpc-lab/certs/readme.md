# Certificados de desenvolvimento

Este diretório é usado apenas para certificados locais de teste (autoassinados).
Para gerar novamente:

```bash
openssl req -x509 -newkey rsa:4096 -nodes -days 365 \
  -keyout server.key -out server.crt -subj "/CN=localhost"