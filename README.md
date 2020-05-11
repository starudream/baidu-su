# Baidu-Su

![Go](https://github.com/starudream/baidu-su/workflows/Go/badge.svg)
![Docker](https://github.com/starudream/baidu-su/workflows/Docker/badge.svg)
![License](https://img.shields.io/badge/License-Apache%20License%202.0-blue)

## Config

```json
{
  "access_key": "",
  "secret_key": "",
  "cron": "* 4 * * *",
  "timezone": "Asia/Shanghai",
  "certs": [
    {
      "domain": "",
      "name": "",
      "crt_path": "",
      "key_path": ""
    }
  ]
}
```

## Usage

![Version](https://img.shields.io/docker/v/starudream/baidu-su)
![Size](https://img.shields.io/docker/image-size/starudream/baidu-su/latest)
![Pull](https://img.shields.io/docker/pulls/starudream/baidu-su)

```bash
docker pull starudream/baidu-su
```

```bash
docker run -d \
    --name baidu-su \
    --restart always \
    -e DEBUG=true \
    -v /opt/docker/baidu-su/config.json:/config.json \
    starudream/baidu-su:latest
```

## License

[Apache License 2.0](./LICENSE)
