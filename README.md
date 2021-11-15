# Baidu-Su

![Golang](https://img.shields.io/github/workflow/status/starudream/baidu-su/Golang/master?style=for-the-badge)
![Docker](https://img.shields.io/github/workflow/status/starudream/baidu-su/Docker/master?style=for-the-badge)
![License](https://img.shields.io/badge/License-Apache%20License%202.0-blue?style=for-the-badge)

## Config

```json
{
    "access_key": "",
    "secret_key": "",
    "bduss": "",
    "cron": "* * 4 * * *",
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

![Version](https://img.shields.io/docker/v/starudream/baidu-su?style=for-the-badge)
![Size](https://img.shields.io/docker/image-size/starudream/baidu-su/latest?style=for-the-badge)
![Pull](https://img.shields.io/docker/pulls/starudream/baidu-su?style=for-the-badge)

```bash
docker pull starudream/baidu-su
```

```bash
docker run -d \
    --name baidu-su \
    --restart always \
    -v /opt/docker/baidu-su/config.json:/config.json \
    -v /usr/local/openresty/nginx/conf/ssl:/ssl:ro \
    starudream/baidu-su:latest
```

```json
{
    "bduss": "xxx",
    "cron": "0 0 1 * * 1",
    "certs": [
        {
            "domain": "52xckl.cn",
            "name": "52xckl.cn",
            "crt_path": "/ssl/52xckl.cn.crt",
            "key_path": "/ssl/52xckl.cn.key"
        }
    ]
}
```

## License

[Apache License 2.0](./LICENSE)
