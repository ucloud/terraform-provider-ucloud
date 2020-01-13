---
layout: "ucloud"
page_title: "UCloud: redis_instance_type"
sidebar_current: "docs-ucloud-appendix-redis-instance-type"
description: |-
   The instance type of redis instance.
---

# Redis Instance Type

The type of redis instance.

## Active-Standby and Distributed

- Introduction: UCloud Redis provides two architectures: Active-Standby Redis and Distributed Redis. Based on the highly reliable dual-machine hot standby architecture and the cluster architecture that can be smoothly extended, it can meet the business requirements of high read-write performance scenarios and elastic expansion and contraction capacity.
- Memory (unit GB): Support 1, 2, 4, 6, 8, 12, 16, 24, 32; The distributed version supports 16 to 1000 and must be divisible by 4.

<table><tr><th colspan="1">Category</th><th colspan="2">Active-Standby</th><th colspan="2">Distributed</th></tr><tr><th rowspan="18">Redis</th><th>InstanceType</th><th>Memory</th><th>InstanceType</th><th>Memory</th></tr><tr><td>redis-master-1</td><td>1</td><td>redis-distributed-16</td><td>16</td> </tr><tr><td>redis-master-2</td><td>2</td><td>redis-distributed-20</td><td>20</td> </tr><tr><td>redis-master-4</td><td>4</td><td>redis-distributed-24</td><td>24</td> </tr><tr><td>redis-master-6</td><td>6</td><td>redis-distributed-28</td><td>28</td> </tr><tr><td>redis-master-8</td><td>8</td><td>...</td><td>...</td> </tr><tr><td>redis-master-12</td><td>12</td><td>redis-distributed-996</td><td>996</td> </tr><tr><td>redis-master-16</td><td>16</td><td>redis-distributed-1000</td><td>1000</td> </tr><tr><td>redis-master-24</td><td>24</td></tr><tr><td>redis-master-32</td><td>32</td> </tr></table>