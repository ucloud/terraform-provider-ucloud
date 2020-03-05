---
layout: "ucloud"
page_title: "UCloud: instance_type"
sidebar_current: "docs-ucloud-appendix-instance-type"
description: |-
   The instance type of the instance.
---

# Instance Type

The type of the instance.

## Normal

- Introduction: Provide the most flexible and free combination of CPU, memory and disk. Suitable for computing, storage, network and other balanced scenarios.

- CPU platform support: Intel IvyBridge/Haswell/Broadwell/Skylake

- CPU Memory combination (support ratio: 2:1-1:12)

- Unit: CPU-kernel Memory-GB

- Range of CPU: 1-32, Range of memory: 1-128

<table><tr><th colspan="1">Category</th><th colspan="3">High CPU（1:1）</th><th colspan="3"> Basic（1:2）</th><th colspan="3"> Standard（1:4）</th><th colspan="3"> High Memory（1:8）</th><th colspan="3"> Customized（2:1-1:12）</th></tr><tr><th rowspan="18">Normal (N) </th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th></tr><tr><td>n-highcpu-1</td><td>1</td><td>1</td><td>n-basic-1</td><td>1</td><td>2</td><td>n-standard-1</td><td>1</td><td>4</td> <td>n-highmem-1</td><td>1</td><td>8</td><td>n-customized-2-1</td><td>2</td><td>1</td></tr><tr><td>n-highcpu-2</td><td>2</td><td>2</td><td>n-basic-2</td><td>2</td><td>4</td><td>n-standard-2</td><td>2</td><td>8</td> <td>n-highmem-2</td><td>2</td><td>16</td><td>n-customized-2-14</td><td>2</td><td>14</td> </tr><tr><td>n-highcpu-4</td><td>4</td><td>4</td><td>n-basic-4</td><td>4</td><td>8</td><td>n-standard-4</td><td>4</td><td>16</td> <td>n-highmem-4</td><td>4</td><td>32</td> <td>n-customized-4-18</td><td>4</td><td>18</td></tr><tr><td>n-highcpu-6</td><td>6</td><td>6</td><td>n-basic-6</td><td>6</td><td>12</td><td>n-standard-6</td><td>6</td><td>24</td> <td>n-highmem-6</td><td>6</td><td>48</td><td>...</td><td>...</td><td>...</td></tr> <tr><td>n-highcpu-8</td><td>8</td><td>8</td><td>n-basic-8</td><td>8</td><td>16</td><td>n-standard-8</td><td>8</td><td>32</td> <td>n-highmem-8</td><td>8</td><td>64</td><td>n-customized-4-48</td><td>4</td><td>48</td> </tr> <tr><td>n-highcpu-10</td><td>10</td><td>10</td><td>n-basic-10</td><td>10</td><td>20</td><td>n-standard-10</td><td>10</td><td>40</td> <td>n-highmem-10</td><td>10</td><td>80</td><td>...</td><td>...</td><td>...</td> </tr> <tr><td>n-highcpu-12</td><td>12</td><td>12</td><td>n-basic-12</td><td>12</td><td>24</td><td>n-standard-12</td><td>12</td><td>48</td> <td>n-highmem-12</td><td>12</td><td>96</td> </tr> <tr><td>n-highcpu-14</td><td>14</td><td>14</td><td>n-basic-14</td><td>14</td><td>28</td><td>n-standard-14</td><td>14</td><td>56</td> <td>n-highmem-14</td><td>14</td><td>112</td> </tr> <tr><td>n-highcpu-16</td><td>16</td><td>16</td><td>n-basic-16</td><td>16</td><td>32</td><td>n-standard-16</td><td>16</td><td>64</td> <td>n-highmem-16</td><td>16</td><td>128</td> </tr> <tr><td>n-highcpu-18</td><td>18</td><td>18</td><td>n-basic-18</td><td>18</td><td>36</td><td>n-standard-18</td><td>18</td><td>72</td></tr> <tr><td>n-highcpu-20</td><td>20</td><td>20</td><td>n-basic-20</td><td>20</td><td>40</td><td>n-standard-20</td><td>20</td><td>80</td></tr> <tr><td>n-highcpu-22</td><td>22</td><td>22</td><td>n-basic-22</td><td>22</td><td>44</td><td>n-standard-22</td><td>22</td><td>88</td></tr> <tr><td>n-highcpu-24</td><td>24</td><td>24</td><td>n-basic-24</td><td>24</td><td>48</td><td>n-standard-24</td><td>24</td><td>96</td></tr> <tr><td>n-highcpu-26</td><td>26</td><td>26</td><td>n-basic-26</td><td>26</td><td>52</td><td>n-standard-26</td><td>26</td><td>104</td></tr> <tr><td>n-highcpu-28</td><td>28</td><td>28</td><td>n-basic-28</td><td>28</td><td>56</td><td>n-standard-28</td><td>28</td><td>112</td></tr> <tr><td>n-highcpu-30</td><td>30</td><td>30</td><td>n-basic-30</td><td>30</td><td>60</td><td>n-standard-30</td><td>30</td><td>120</td></tr> <tr><td>n-highcpu-32</td><td>32</td><td>32</td><td>n-basic-32</td><td>32</td><td>64</td><td>n-standard-32</td><td>32</td><td>128</td></tr> </table>

## OutStanding (public beta)

- Introduction: The latest generation of cloud hosts with excellent computing, storage and network performance. Suitable for the overall requirements scenario.

- CPU Platform Support: Intel Skylake/Cascadelake

- CPU Memory Combination (support ratio: 1:1-1:8)

- Unit: CPU-kernel Memory-GB

- Range of CPU: 4-64, Range of memory: 4-256

- Limit: 
    - Currently only supports the `cn-bj2-05`(`availability_zone`) in `cn-bj2`(`region`)
    - Must set `boot_disk_type` to `cloud_ssd`
    - Can only use specified Image (image type is `base` and the name of which is prefix with "高内核")
    - Can only attach specified Disk（the disk attached to instance must be `rssd_data_disk` (RDMA-SSD) cloud disk if required)

<table><tr><th colspan="1">Category</th><th colspan="3">High CPU（1:1）</th><th colspan="3"> Basic（1:2）</th><th colspan="3"> Standard（1:4）</th><th colspan="3"> High Memory（1:8）</th></tr><tr><th rowspan="6">OutStanding (O) </th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th></tr><tr><td>o-highcpu-4</td><td>4</td><td>4</td><td>o-basic-4</td><td>4</td><td>8</td><td>o-standard-4</td><td>4</td><td>16</td> <td>o-highmem-4</td><td>4</td><td>32</td> </tr><tr><td>o-highcpu-8</td><td>8</td><td>8</td><td>o-basic-8</td><td>8</td><td>16</td><td>o-standard-8</td><td>8</td><td>32</td> <td>o-highmem-8</td><td>8</td><td>64</td> </tr> <tr><td>o-highcpu-16</td><td>16</td><td>16</td><td>o-basic-16</td><td>16</td><td>32</td><td>o-standard-16</td><td>16</td><td>64</td> <td>o-highmem-16</td><td>16</td><td>128</td> </tr> <tr><td>o-highcpu-32</td><td>32</td><td>32</td><td>o-basic-32</td><td>32</td><td>64</td><td>o-standard-32</td><td>32</td><td>128</td><td>o-highmem-32</td><td>32</td><td>256</td></tr> <tr><td>o-highcpu-64</td><td>64</td><td>64</td><td>o-basic-64</td><td>64</td><td>128</td><td>o-standard-64</td><td>64</td><td>256</td></tr> </table>

## High Frequency (C)

- Introduction: Models with a CPU frequency of 3.0 GHz or higher are suitable for computing services such as high-frequency trading, rendering, artificial intelligence, etc.

- CPU Platform Support: Intel Skylake

- CPU Memory Combination (support ratio: 1:1-1:8)

- Unit: CPU-kernel Memory-GB

- Range of CPU: 1-32, Range of memory: 1-128

- Limit: 
    - Must set `boot_disk_type` to `cloud_ssd`

<table><tr><th colspan="1">Category</th><th colspan="3">High CPU（1:1）</th><th colspan="3">Basic（1:2）</th><th colspan="3">Standard（1:4）</th><th colspan="3">High Memory（1:8）</th></tr><tr><th rowspan="8">High Frequency (C)</th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th><th>InstanceType</th><th>CPU</th><th>Memory</th></tr><tr><td>n-highcpu-1</td><td>1</td><td>1</td><td>n-basic-1</td><td>1</td><td>2</td><td>n-standard-1</td><td>1</td><td>4</td><td>n-highmem-1</td><td>1</td><td>8</td><tr><td>n-highcpu-2</td><td>2</td><td>2</td><td>n-basic-2</td><td>2</td><td>4</td><td>n-standard-2</td><td>2</td><td>8</td><td>n-highmem-2</td><td>2</td><td>16</td><tr><tr><td>o-highcpu-4</td><td>4</td><td>4</td><td>o-basic-4</td><td>4</td><td>8</td><td>o-standard-4</td><td>4</td><td>16</td><td>o-highmem-4</td><td>4</td><td>32</td></tr><tr><td>n-highcpu-8</td><td>8</td><td>8</td><td>n-basic-8</td><td>8</td><td>16</td><td>n-standard-8</td><td>8</td><td>32</td><td>n-highmem-8</td><td>8</td><td>64</td><tr><td>n-highcpu-16</td><td>16</td><td>16</td><td>n-basic-16</td><td>16</td><td>32</td><td>n-standard-16</td><td>16</td><td>64</td><td>n-highmem-16</td><td>16</td><td>128</td></tr><tr><td>n-highcpu-32</td><td>32</td><td>32</td><td>n-basic-32</td><td>32</td><td>64</td><td>n-standard-32</td><td>32</td><td>128</td></tr></table>