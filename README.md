# RouterOS REST API bridge

This project aims to provide REST capability to Mikrotik RouterOS devices running pre-[RouterOS v7.1beta4](https://help.mikrotik.com/docs/spaces/ROS/pages/47579162/REST+API).
Only requirement for device is to have [API enabled](https://help.mikrotik.com/docs/spaces/ROS/pages/47579160/API).

### Example config:
```yaml
---
devices:
  rb941:
    username: admin
    password: admin
    tls:
      verify: false
aliases:
  arp:
    path: /ip/arp
```

### Example usage of REST client:

This invocation of `curl` will read ARP table from device
```shell
curl http://localhost:22003/api/v1/data/rb941/arp
```

#### Output (truncated)
```json
[
  {
    ".id": "*1",
    "DHCP": "false",
    "address": "192.168.30.31",
    "complete": "true",
    "disabled": "false",
    "dynamic": "true",
    "interface": "bridge1",
    "invalid": "false",
    "mac-address": "RE:DA:CT:ED:00:12",
    "published": "false"
  },
...
]
```

By design, any mutable operation (create/udpate/delete) is disabled on alias/path level.
To enable it, use alias config:

```yaml
aliases:
  arp:
    path: /ip/arp
    create: true
    delete: true
    update: true
```
