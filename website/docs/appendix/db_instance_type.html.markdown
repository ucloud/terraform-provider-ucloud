---
subcategory: "Appendix"
layout: "ucloud"
page_title: "UCloud: db_instance_type"
description: |-
   The instance type of DB instance.
---

# DB Instance Type

The type of DB instance.

## Highly Availability SATA SSD

- Introduction: The high-availability SATA SSD version use the dual main hot standby structure which suitable for business scenarios that require high database performance.
- Memory: Supports 1, 2, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192, 256, 320 (unit GB)

<table><tr><th colspan="1">Category</th><th colspan="2">Mysql</th><th colspan="2">Percona</th></tr><tr><th rowspan="18">Mysql/Percona DB</th><th>InstanceType</th><th>Memory</th><th>InstanceType</th><th>Memory</th></tr><tr><td>mysql-ha-1</td><td>1</td><td>percona-ha-1</td><td>1</td> </tr><tr><td>mysql-ha-2</td><td>2</td><td>percona-ha-2</td><td>2</td> </tr><tr><td>mysql-ha-4</td><td>4</td><td>percona-ha-4</td><td>4</td> </tr><tr><td>mysql-ha-6</td><td>6</td><td>percona-ha-6</td><td>6</td> </tr><tr><td>mysql-ha-8</td><td>8</td><td>percona-ha-8</td><td>8</td> </tr><tr><td>mysql-ha-12</td><td>12</td><td>percona-ha-12</td><td>12</td> </tr><tr><td>mysql-ha-16</td><td>16</td><td>percona-ha-16</td><td>16</td> </tr><tr><td>mysql-ha-24</td><td>24</td><td>percona-ha-24</td><td>24</td> </tr><tr><td>mysql-ha-32</td><td>32</td><td>percona-ha-32</td><td>32</td> </tr><tr><td>mysql-ha-48</td><td>48</td><td>percona-ha-48</td><td>48</td> </tr><tr><td>mysql-ha-64</td><td>64</td><td>percona-ha-64</td><td>64</td> </tr><tr><td>mysql-ha-96</td><td>96</td><td>percona-ha-96</td><td>96</td> </tr><tr><td>mysql-ha-128</td><td>128</td><td>percona-ha-128</td><td>128</td> </tr><tr><td>mysql-ha-192</td><td>192</td><td>percona-ha-192</td><td>192</td> </tr><tr><td>mysql-ha-256</td><td>256</td><td>percona-ha-256</td><td>256</td> </tr><tr><td>mysql-ha-320</td><td>320</td><td>percona-ha-320</td><td>320</td> </tr></table>

## Highly Availability NVMe SSD (public beta)

- Introduction: The high-availability NVMe SSD version use the dual main hot standby structure which is new generation ultra high performance cloud disk products, suitable for business scenarios with high capacity and low latency requirements.
- Memory: Supports 2, 4, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192, 256, 320 (unit GB)
- Limit: 
    - Currently not fully support by all `availability_zone`in the `region`, please proceed to [UCloud console](https://console.ucloud.cn/uhost/uhost/create) for more details.
    - Currently not fully support by all `engine` and `engine_version`, we know that it supports `engine`: `mysql`; `engine_version`: `5.6`, `5.7`; please proceed to [UCloud console](https://console.ucloud.cn/uhost/uhost/create) for more details.
    

<table><tr> <th colspan="1">Category</th> <th colspan="2">Mysql</th></tr><tr> <th rowspan="16">Mysql DB</th> <th>InstanceType</th> <th>Memory</th></tr><tr> <td>mysql-ha-nvme-2</td> <td>2</td></tr><tr> <td>mysql-ha-nvme-4</td> <td>4</td></tr><tr> <td>mysql-ha-nvme-8</td> <td>8</td></tr><tr> <td>mysql-ha-nvme-12</td> <td>12</td></tr><tr> <td>mysql-ha-nvme-16</td> <td>16</td></tr><tr> <td>mysql-ha-nvme-24</td> <td>24</td></tr><tr> <td>mysql-ha-nvme-32</td> <td>32</td></tr><tr> <td>mysql-ha-nvme-48</td> <td>48</td></tr><tr> <td>mysql-ha-nvme-64</td> <td>64</td></tr><tr> <td>mysql-ha-nvme-96</td> <td>96</td></tr><tr> <td>mysql-ha-nvme-128</td> <td>128</td></tr><tr> <td>mysql-ha-nvme-192</td> <td>192</td></tr><tr> <td>mysql-ha-nvme-256</td> <td>256</td></tr><tr> <td>mysql-ha-nvme-320</td> <td>320</td></tr></table>