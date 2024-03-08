## v1.38.7 (2024-03-08)

FEATURES:

* `resource/ucloud_disk_attachment`: add `device_name` attribute.

## v1.38.6 (2024-02-26)

BUG FIXES:

* `resource/ucloud_instance`: ignore min_cpu_platform change from null to Intel/Auto

## v1.38.5 (2024-02-26)

BUG FIXES:

* `resource/ucloud_instance`: add a default value of min_cpu_platform

## v1.38.4 (2024-02-01)

BUG FIXES:

* `resource/ucloud_instance`: fix panic issue while detaching disk in a host

## 1.38.3 (2023-11-22)

BUG FIXES:

* `datasource/ucloud_iam_policy_document`: fix data type of version.

## 1.38.2 (2023-09-27)

FEATURES:

* add `reboot_instance_for_resizing` attribute to `ucloud_disk` resource

## 1.38.1 (2023-09-03)

FEATURES:

* add rdma_cluster_id attribute to bare metal instance

## 1.38.0 (2023-09-03)

FEATURES:

* **New Resource:** `ucloud_baremetal_instance`.
* **New Datasource:** `ucloud_baremetal_images`.

## 1.37.0 (2023-08-27)

FEATURES:

* **New Resource:** `ucloud_instance_state`.
* feat: support stop_instance_before_detaching in disk_attachment resource.

## 1.36.1 (2023-05-12)

DOCS:

* fix documentation


## 1.36.0 (2023-04-17)

FEATURES:

* **New Resource:** `ucloud_iam_access_key`.
* **New Resource:** `ucloud_iam_group`.
* **New Resource:** `ucloud_iam_user`.
* **New Resource:** `ucloud_iam_group_membership`.
* **New Resource:** `ucloud_iam_project`.
* **New Resource:** `ucloud_iam_policy`.
* **New Resource:** `ucloud_iam_user_policy_attachment`.
* **New Resource:** `ucloud_iam_group_policy_attachment`.
* **New Datasource:** `ucloud_iam_groups`.
* **New Datasource:** `ucloud_iam_users`.
* **New Datasource:** `ucloud_iam_projects`.
* **New Datasource:** `ucloud_iam_policy`.
* **New Datasource:** `ucloud_iam_policy_document`.

## 1.35.1 (2023-03-21)

DOCS:

* update documentation

## 1.35.0 (2023-03-20)

FEATURES:

* **New Resource:** `ucloud_anti_ddos_instance`.
* **New Resource:** `ucloud_anti_ddos_ip`.
* **New Resource:** `ucloud_anti_ddos_allowed_domain`.
* **New Resource:** `ucloud_anti_ddos_rule`.
* **New Datasource:** `ucloud_anti_ddos_instances`.
* **New Datasource:** `ucloud_anti_ddos_ips`.

## 1.34.1 (2023-02-27)

BUG FIXES:

* resource/uk8s_node: fix disk type validation.
* resource/uk8s_cluster: fix disk type validation.

## 1.34.0 (2023-01-17)

FEATURES:

* feat: support create UDB from backup
* feat: datasource to query UDB backups
* chore: remove cube

## 1.33.0 (2022-11-30)

FEATURES:

* feat: support custom image for UK8S master nodes.

## 1.32.5 (2022-11-07)

BUG FIXES:

* resource/uk8s_node: use correct `ImageId` param.

## 1.32.4 (2022-10-11)

BUG FIXES:

* resource/ucloud_redis_instance: still validate version after apply.

## 1.32.3 (2022-10-10)

FEATURES:

* resource/ucloud_redis_instance: remove version validation for field `engine_version`.

## 1.32.2 (2022-09-20)

NOTES:

* doc: update desc for `rdma_cluster_id` in `resource/ucloud_disk`.

## 1.32.1 (2022-09-19)

NOTES:

* fix `rdma_cluster_id` doc.

## 1.32.0 (2022-09-19)

FEATURES:

* resource/ucloud_instance, resources/ucloud_disk: add field `rdma_cluster_id` to fix rdma missing issue. ([#127](https://github.com/ucloud/terraform-provider-ucloud/issues/127))

## 1.31.3 (2022-05-12)

ENHANCEMENTS:

* ci: add test workflow to github actions ([#120](https://github.com/ucloud/terraform-provider-ucloud/pull/120))

BUG FIXES:

* resource/ucloud_lb_listener: fix lb listener validation. ([#117](https://github.com/ucloud/terraform-provider-ucloud/pull/117))

## 1.31.2 (2022-04-27)

BUG FIXES:

* build: fix build failure when using go 1.18 on darwin_arm64. ([#115](https://github.com/ucloud/terraform-provider-ucloud/issues/115))

## 1.31.1 (2022-04-01)

BUG FIXES:

* resource/ucloud_instance: remove `cpu_platform` validator. ([#114](https://github.com/ucloud/terraform-provider-ucloud/issues/114))

## 1.31.0 (2022-01-26)

ENHANCEMENTS:

* resource/ucloud_eip_association: add `resource_type` to argument. ([#110](https://github.com/ucloud/terraform-provider-ucloud/issues/110))
* resource/ucloud_lb_attachment: add `resource_type` to argument. ([#110](https://github.com/ucloud/terraform-provider-ucloud/issues/110))


## 1.30.0 (2021-10-26)

FEATURES:

* **New Resource:** `ucloud_ufs_volume_mount_point`. ([#109](https://github.com/ucloud/terraform-provider-ucloud/issues/109))

ENHANCEMENTS:

* resource/ucloud_uk8s_cluster: add custom diff precheck of `boot_disk_type` and `instance_type` about outstanding machine. ([#109](https://github.com/ucloud/terraform-provider-ucloud/issues/109))
* resource/ucloud_uk8s_node: add custom diff precheck of `boot_disk_type` and `instance_type` about outstanding machine. ([#109](https://github.com/ucloud/terraform-provider-ucloud/issues/109))

BUG FIXES:

* resource/ucloud_uk8s_cluster: fix the default value to take effect of `boot_disk_type`, `data_disk_type` and `min_cpu_platform`. ([#109](https://github.com/ucloud/terraform-provider-ucloud/issues/109))

## 1.29.0 (2021-06-22)

FEATURES:

* **New Resource:** `ucloud_uk8s_cluster`. ([#103](https://github.com/ucloud/terraform-provider-ucloud/issues/103))
* **New Resource:** `ucloud_uk8s_node`. ([#103](https://github.com/ucloud/terraform-provider-ucloud/issues/103))

## 1.28.0 (2021-06-03)

ENHANCEMENTS:

* resource/ucloud_cube_pod: add `security_group` to argument.([#102](https://github.com/ucloud/terraform-provider-ucloud/issues/102))

BUG FIXES:

* resource/ucloud_cube_pod: fix the [issue](https://github.com/ucloud/terraform-provider-ucloud/issues/101) .([#102](https://github.com/ucloud/terraform-provider-ucloud/issues/102))

## 1.27.0 (2021-05-18)

FEATURES:

* **New Datasource:** `ucloud_ufs_volumes`. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))
* **New Datasource:** `ucloud_us3_buckets`. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))
* **New Resource:** `ucloud_ufs_volume`. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))
* **New Resource:** `ucloud_us3_bucket`. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))
* **New Resource:** `ucloud_cube_pod`. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))

ENHANCEMENTS:

* datasource/ucloud_db_parameter_groups: add `class_type` to argument. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))
* resource/ucloud_db_instance: add `postgresql` enum value to `engine`. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))
* resource/ucloud_instance: add `Ampere/Altra` enum value to `min_cpu_platform`. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))
* resource/ucloud_nat_gateway: support update `security_group`. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))
* resource/ucloud_redis_instance: support arguments about backup Redis instance and `standby_zone` only for Active-Standby Redis. ([#95](https://github.com/ucloud/terraform-provider-ucloud/issues/95))

## 1.26.0 (2021-04-28)

ENHANCEMENTS:

* Update for go Version 1.16. ([#92](https://github.com/ucloud/terraform-provider-ucloud/issues/92))
* the release support `darwin_arm64`. ([#92](https://github.com/ucloud/terraform-provider-ucloud/issues/92))
* resource/ucloud_redis_instance: add the version `6.0` of the `engine_version` supported. ([#93](https://github.com/ucloud/terraform-provider-ucloud/issues/93))
* resource/ucloud_disk: open the usage restrictions of `rssd_data_disk`. ([#93](https://github.com/ucloud/terraform-provider-ucloud/issues/93))

## 1.25.1 (2021-04-08)

BUG FIXES:

* resource/ucloud_lb: Fix the validation about `internal` and `subnet_id`. ([#90](https://github.com/ucloud/terraform-provider-ucloud/issues/90))
* resource/ucloud_redis_instance: Fix the validation about `engine_version`.([#90](https://github.com/ucloud/terraform-provider-ucloud/issues/90))

## 1.25.0 (2020-12-11)

ENHANCEMENTS:

* resource/ucloud_lb: add `listen_type` to the argument. ([#88](https://github.com/ucloud/terraform-provider-ucloud/issues/88))
* datasource/ucloud_subnets: add `vpc_id` to the attributes. ([#88](https://github.com/ucloud/terraform-provider-ucloud/issues/88))

## 1.24.1 (2020-11-06)

BUG FIXES:

* resource/ucloud_db_instance: Fix the error about updating only one of `instance_storage` and `instance_type`. ([#86](https://github.com/ucloud/terraform-provider-ucloud/issues/86))

## 1.24.0 (2020-10-23)

ENHANCEMENTS:

* resource/ucloud_instance:  add `network_interface` to the argument in order to create and associate `eip` with the instance.([#85](https://github.com/ucloud/terraform-provider-ucloud/issues/85))

## 1.23.0 (2020-09-07)

ENHANCEMENTS:

* resource/ucloud_vpc_peering_connection: add `peer_region` to argument in order to support vpc intercommunication in different regions(([#83](https://github.com/ucloud/terraform-provider-ucloud/issues/83)))

## 1.22.0 (2020-08-27)

ENHANCEMENTS:

* resource/ucloud_db_instance: update `instance_type` about NVMe SSD DB [#69](https://github.com/ucloud/terraform-provider-ucloud/issues/69)
* resource/ucloud_instance: update `instance_type` about OS instance [#69](https://github.com/ucloud/terraform-provider-ucloud/issues/69)

## 1.21.0 (July 27, 2020)

ENHANCEMENTS:

* Update for Terraform 0.13
* Provider was moved to https://github.com/ucloud/terraform-provider-ucloud

## 1.20.0 (June 08, 2020)

ENHANCEMENTS:

* resource/ucloud_instance: add `min_cpu_platform` to argument([#68](https://github.com/terraform-providers/terraform-provider-ucloud/issues/68))
* resource/ucloud_instance: add `cpu_platform` to attributes([#68](https://github.com/terraform-providers/terraform-provider-ucloud/issues/68))


## 1.19.0 (April 15, 2020)

FEATURES:

* **New Datasource:** `ucloud_db_parameter_groups` ([#66](https://github.com/terraform-providers/terraform-provider-ucloud/issues/66))

ENHANCEMENTS:

* resource/ucloud_db_instance: add `parameter_group` to argument ([#66](https://github.com/terraform-providers/terraform-provider-ucloud/issues/66))
* resource/ucloud_db_instance: add `timeouts` to argument ([#66](https://github.com/terraform-providers/terraform-provider-ucloud/issues/66))
* resource/ucloud_instance: add `timeouts` to argument ([#66](https://github.com/terraform-providers/terraform-provider-ucloud/issues/66))

## 1.18.0 (April 10, 2020)

ENHANCEMENTS:

* resource/ucloud_instance: add `data_disks` to argument ([#65](https://github.com/terraform-providers/terraform-provider-ucloud/issues/65))

## 1.17.0 (March 31, 2020)

ENHANCEMENTS:

* resource/ucloud_instance: add `user_data` to argument([#64](https://github.com/terraform-providers/terraform-provider-ucloud/issues/64))
* resource/ucloud_lb: add `security_group` to argument([#64](https://github.com/terraform-providers/terraform-provider-ucloud/issues/64))
* resource/ucloud_redis_instance: add `5.0` enums of `engine_version` and refresh the `instance_type`([#64](https://github.com/terraform-providers/terraform-provider-ucloud/issues/64))

## 1.16.0 (March 06, 2020)

ENHANCEMENTS:

* resource/ucloud_instance: add high frequency(c) instance type ([#61](https://github.com/terraform-providers/terraform-provider-ucloud/issues/61))
* datasource/ucloud_images: add argument `ids`([#61](https://github.com/terraform-providers/terraform-provider-ucloud/issues/61))
* datasource support `ids` as attributes ([#61](https://github.com/terraform-providers/terraform-provider-ucloud/issues/61))

## 1.15.1 (January 13, 2020)

BUG FIXES:

* resource/ucloud_instance: Fix the validate of `instance_type`([#58](https://github.com/terraform-providers/terraform-provider-ucloud/issues/58))

## 1.15.0 (December 27, 2019)

NOTES:

* provider: The underlying Terraform codebase dependency for the provider SDK and acceptance testing framework has been migrated from `github.com/hashicorp/terraform` to `github.com/hashicorp/terraform-plugin-sdk`. They are functionality equivalent and this should only impact codebase development to switch imports. For more information see the [Terraform Plugin SDK page in the Extending Terraform documentation](https://www.terraform.io/docs/extend/plugin-sdk.html)([#54](https://github.com/terraform-providers/terraform-provider-ucloud/issues/54))

ENHANCEMENTS:

* resource/ucloud_instance: add `private_ip` to argument([#56](https://github.com/terraform-providers/terraform-provider-ucloud/issues/56))

## 1.14.1 (October 18, 2019)

BUG FIXES:

* provider: Fix the provider about `insecure` didn't take effect([#53](https://github.com/terraform-providers/terraform-provider-ucloud/issues/53))

## 1.14.0 (October 13, 2019)

FEATURES:

* **New Datasource:** `ucloud_nat_gateways`([#51](https://github.com/terraform-providers/terraform-provider-ucloud/issues/51))
* **New Datasource:** `ucloud_vpn_gateways`([#51](https://github.com/terraform-providers/terraform-provider-ucloud/issues/51))
* **New Datasource:** `ucloud_vpn_customer_gateways`([#51](https://github.com/terraform-providers/terraform-provider-ucloud/issues/51))
* **New Datasource:** `ucloud_vpn_connections`([#51](https://github.com/terraform-providers/terraform-provider-ucloud/issues/51))

ENHANCEMENTS:

* resource/ucloud_nat_gateway: enable to import([#51](https://github.com/terraform-providers/terraform-provider-ucloud/issues/51))
* resource/ucloud_vpn_gateway: enable to import([#51](https://github.com/terraform-providers/terraform-provider-ucloud/issues/51))
* resource/ucloud_vpn_customer_gateway: enable to import([#51](https://github.com/terraform-providers/terraform-provider-ucloud/issues/51))
* resource/ucloud_vpn_connection: enable to import([#51](https://github.com/terraform-providers/terraform-provider-ucloud/issues/51))

BUG FIXES:

* resource/ucloud_subnet: Fix the validate of `cidr_block`([#50](https://github.com/terraform-providers/terraform-provider-ucloud/issues/50))
* resource/ucloud_vpc: Fix the validate of `cidr_blocks`([#50](https://github.com/terraform-providers/terraform-provider-ucloud/issues/50))
* resource/ucloud_vpn_connection: Fix the validate of `remote_subnets`([#50](https://github.com/terraform-providers/terraform-provider-ucloud/issues/50))

## 1.13.0 (September 30, 2019)

FEATURES:

* **New Resource:** `ucloud_nat_gateway`([#48](https://github.com/terraform-providers/terraform-provider-ucloud/issues/48))
* **New Resource:** `ucloud_nat_gateway_rule`([#48](https://github.com/terraform-providers/terraform-provider-ucloud/issues/48))
* **New Resource:** `ucloud_vpn_gateway`[[#48](https://github.com/terraform-providers/terraform-provider-ucloud/issues/48)]
* **New Resource:** `ucloud_vpn_customer_gateway`([#48](https://github.com/terraform-providers/terraform-provider-ucloud/issues/48))
* **New Resource:** `ucloud_vpn_connection`([#48](https://github.com/terraform-providers/terraform-provider-ucloud/issues/48))
* **New Resource:** `ucloud_vip`([#48](https://github.com/terraform-providers/terraform-provider-ucloud/issues/48))

ENHANCEMENTS:

* resource/ucloud_instance: Add argument `allow_stopping_for_update`([#48](https://github.com/terraform-providers/terraform-provider-ucloud/issues/48))
* resource/ucloud_db_instance: Update the range of `instance_type` and `instance_storage`([#48](https://github.com/terraform-providers/terraform-provider-ucloud/issues/48))

BUG FIXES:

* resource/ucloud_subnet: Fix the problem about the subnet can not be deleted while associated resources of it have been deleted.([#48](https://github.com/terraform-providers/terraform-provider-ucloud/issues/48))

## 1.12.1 (August 26, 2019)

BUG FIXES:
* Fix the `duration` about it can specify zero-value distinguish from null-value([#45](https://github.com/terraform-providers/terraform-provider-ucloud/issues/45))
* resource/ucloud_db_instance: Fix the `backup_begin_time` about it can specify zero-value distinguish from null-value([#45](https://github.com/terraform-providers/terraform-provider-ucloud/issues/45))
* resource/ucloud_instance: Fix the content of return err about `cloud_normal` of `boot_disk_type` is not supported currently([#45](https://github.com/terraform-providers/terraform-provider-ucloud/issues/45))
* resource/ucloud_lb_listener: Fix the `idle_timeout` about it can specify zero-value distinguish from null-value([#45](https://github.com/terraform-providers/terraform-provider-ucloud/issues/45))
* resource/ucloud_lb_ssl_attachment: Fix the inaccurate plan about creating lb ssl attachment([#46](https://github.com/terraform-providers/terraform-provider-ucloud/issues/46))
* resource/ucloud_lb_ssl_attachment: Fix the api err about deleting lb ssl attachment([#46](https://github.com/terraform-providers/terraform-provider-ucloud/issues/46))
* resource/ucloud_disk_attachment: Fix the api err about delete disk attachment([#47](https://github.com/terraform-providers/terraform-provider-ucloud/issues/47))

## 1.12.0 (August 23, 2019)

ENHANCEMENTS:

* datasource/ucloud_images: Add argument `most_recent`([#43](https://github.com/terraform-providers/terraform-provider-ucloud/issues/43))
* resource/ucloud_disk: Update customdiff precheck about `disk_size`([#43](https://github.com/terraform-providers/terraform-provider-ucloud/issues/43))

BUG FIXES:

* Fix the validate function of `duration` about allow the value `0` when pay by month([#43](https://github.com/terraform-providers/terraform-provider-ucloud/issues/43))
* resource/ucloud_disk: Fix the problem about resize the `disk_size` when the disk have attached to the instance([#43](https://github.com/terraform-providers/terraform-provider-ucloud/issues/43))

## 1.11.1 (July 19, 2019)

BUG FIXES:

* Fix the problem of WaitForState function about keep waiting until the timeout occurs in some situation [[#41](https://github.com/terraform-providers/terraform-provider-ucloud/issues/41)]

## 1.11.0 (July 09, 2019)
FEATURES:

* **New Resource:** `ucloud_isolation_group` ([#38](https://github.com/terraform-providers/terraform-provider-ucloud/issues/38))

ENHANCEMENTS:

* update terraform SDK to 0.12.2 ([#37](https://github.com/terraform-providers/terraform-provider-ucloud/issues/37))

## 1.10.1 (June 17, 2019)

BUG FIXES:

* resource/ucloud_instance: Fix the problem that the instance updates some specific attributes (`instance_type`, `root_password`, `boot_disk_size`, `data_disk_size`) without automatically starting up ([#36](https://github.com/terraform-providers/terraform-provider-ucloud/issues/36))

## 1.10.0 (June 10, 2019)

FEATURES:

* **Terraform 0.12** update terraform SDK to 0.12 ([#34](https://github.com/terraform-providers/terraform-provider-ucloud/issues/34))

## 1.9.0 (May 21, 2019)

FEATURES:

* **New Resource:** `ucloud_redis_instance` ([#33](https://github.com/terraform-providers/terraform-provider-ucloud/issues/33))
* **New Resource:** `ucloud_memcache_instance` ([#33](https://github.com/terraform-providers/terraform-provider-ucloud/issues/33))

ENHANCEMENTS:

* datasource/ucloud_db_instances: Add attribute `private_ip`([#33](https://github.com/terraform-providers/terraform-provider-ucloud/issues/33))
* resource/ucloud_db_instance: Add attribute `private_ip`([#33](https://github.com/terraform-providers/terraform-provider-ucloud/issues/33))

BUG FIXES:

* resource/ucloud_lb_listener: Fix `port` to ForceNew and fix default value about it ([#32](https://github.com/terraform-providers/terraform-provider-ucloud/issues/32))
* resource/ucloud_lb: Fix `vpc_id` `subnet_id` to ForceNew ([#32](https://github.com/terraform-providers/terraform-provider-ucloud/issues/32))
* resource/ucloud_db_instance: Fix `tag` to ForceNew ([#32](https://github.com/terraform-providers/terraform-provider-ucloud/issues/32))
* resource/ucloud_subnet: Fix `remark` to ForceNew ([#32](https://github.com/terraform-providers/terraform-provider-ucloud/issues/32))
* resource/ucloud_vpc_peering_connection: Fix `peer_project_id` to ForceNew ([#32](https://github.com/terraform-providers/terraform-provider-ucloud/issues/32))
* resource/ucloud_udpn_connection: Fix `charge_type` `duration`to ForceNew ([#32](https://github.com/terraform-providers/terraform-provider-ucloud/issues/32))


## 1.8.0 (May 14, 2019)

ENHANCEMENTS:

* datasource/ucloud_disks: Add attribute `availability_zone`([#30](https://github.com/terraform-providers/terraform-provider-ucloud/issues/30))
* datasource/ucloud_disks: Add possible value `rssd_data_disk` of `disk_type` ([#30](https://github.com/terraform-providers/terraform-provider-ucloud/issues/30))
* resource/ucloud_disk: Add possible value `rssd_data_disk` of `disk_type` ([#30](https://github.com/terraform-providers/terraform-provider-ucloud/issues/30))
* resource/ucloud_instance: Optimization of validate about `instance_type` ([#30](https://github.com/terraform-providers/terraform-provider-ucloud/issues/30))
* resource/ucloud_instance: Add Outstanding `instance_type` ([#30](https://github.com/terraform-providers/terraform-provider-ucloud/issues/30))

## 1.7.0 (May 10, 2019)

ENHANCEMENTS:

* datasource/ucloud_instances: Add attribute `vpc_id `, `subnet_id`, `private_ip` ([#27](https://github.com/terraform-providers/terraform-provider-ucloud/issues/27))
* resource/ucloud_instance: Add attribute `private_ip` ([#27](https://github.com/terraform-providers/terraform-provider-ucloud/issues/27))
* resource/ucloud_lb_listener: Update customdiff precheck about `protocol` and `listen_type` ([#27](https://github.com/terraform-providers/terraform-provider-ucloud/issues/27))

BUG FIXES:

* resource/ucloud_vpc: Fix `cidr_blocks` validate func ([#28](https://github.com/terraform-providers/terraform-provider-ucloud/issues/28))
* resource/ucloud_subnet: Fix `cidr_block` validate func ([#28](https://github.com/terraform-providers/terraform-provider-ucloud/issues/28))
* resource/ucloud_instance: Fix `image_id` read ([#28](https://github.com/terraform-providers/terraform-provider-ucloud/issues/28))

## 1.6.0 (April 12, 2019)

FEATURES:

* **New Datasource:** `ucloud_db_instances`([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))
* **New Datasource:** `ucloud_lb_ssls`([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))
* **New Datasource:** `ucloud_security_groups`([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))
* **New Datasource:** `ucloud_vpcs`([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))
* **New Datasource:** `ucloud_subnets`([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))

ENHANCEMENTS:

* datasource/ucloud_lbs: Add attribute `internal` ([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))
* datasource/ucloud_instances: Add argument `name_regex` ([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))
* datasource/ucloud_eips: Add argument `name_regex` ([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))
* datasource/ucloud_projects: Add argument `name_regex` ([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))
* datasource/ucloud_zones: Add attribute `total_count` ([#24](https://github.com/terraform-providers/terraform-provider-ucloud/issues/24))

## 1.5.0 (April 01, 2019)

FEATURES:

* **New Datasource:** `ucloud_disks`([#23](https://github.com/terraform-providers/terraform-provider-ucloud/issues/23))
* **New Datasource:** `ucloud_lbs`([#23](https://github.com/terraform-providers/terraform-provider-ucloud/issues/23))
* **New Datasource:** `ucloud_lb_listeners`([#23](https://github.com/terraform-providers/terraform-provider-ucloud/issues/23))
* **New Datasource:** `ucloud_lb_rules`([#23](https://github.com/terraform-providers/terraform-provider-ucloud/issues/23))
* **New Datasource:** `ucloud_lb_attachments`([#23](https://github.com/terraform-providers/terraform-provider-ucloud/issues/23))

ENHANCEMENTS:

* resource/ucloud_lb: Deprecated attribute `expire_time` for optimizing outputs ([#23](https://github.com/terraform-providers/terraform-provider-ucloud/issues/23))

## 1.4.0 (March 18, 2019)

ENHANCEMENTS:

* resource/ucloud_db_instance: Shorten the waiting time ([#22](https://github.com/terraform-providers/terraform-provider-ucloud/issues/22))
* resource/ucloud_disk: Shorten the waiting time and update states ([#22](https://github.com/terraform-providers/terraform-provider-ucloud/issues/22))
* resource/ucloud_disk_attachment: Shorten the waiting time ([#22](https://github.com/terraform-providers/terraform-provider-ucloud/issues/22))
* resource/ucloud_eip: Shorten the waiting time and update states ([#22](https://github.com/terraform-providers/terraform-provider-ucloud/issues/22))
* resource/ucloud_lb_listener: Shorten the waiting time and update states ([#22](https://github.com/terraform-providers/terraform-provider-ucloud/issues/22))
* resource/ucloud_lb_attachment: Shorten the waiting time and update states ([#22](https://github.com/terraform-providers/terraform-provider-ucloud/issues/22))

## 1.3.1 (March 15, 2019)

BUG FIXES:

* resource/ucloud_lb_listener: Fix lb listener import ([#21](https://github.com/terraform-providers/terraform-provider-ucloud/issues/21))
* resource/ucloud_lb_attachment: Fix lb attachment import ([#21](https://github.com/terraform-providers/terraform-provider-ucloud/issues/21))
* resource/ucloud_lb_rule: Fix lb rule import ([#21](https://github.com/terraform-providers/terraform-provider-ucloud/issues/21))

## 1.3.0 (March 12, 2019)

ENHANCEMENTS:

* resource/ucloud_db_instance: Add default password ([#18](https://github.com/ucloud/terraform-provider-ucloud/issues/18))
* resource/ucloud_lb: Deprecated `charge_type` ([#18](https://github.com/ucloud/terraform-provider-ucloud/issues/18))

BUG FIXES:

* resource/ucloud_lb: Fix lb import about `charge_type` and `internal` ([#18](https://github.com/ucloud/terraform-provider-ucloud/issues/18))

## 1.2.1 (March 06, 2019)

ENHANCEMENTS:

* resource/ucloud_instance: Add default root password ([#15](https://github.com/terraform-providers/terraform-provider-ucloud/issues/15))

BUG FIXES:

* resource/ucloud_instance: Fix validate cloud disk import ([#15](https://github.com/terraform-providers/terraform-provider-ucloud/issues/15))

## 1.2.0 (March 05, 2019)

FEATURES:

* **New Resource:** `ucloud_db_instance` ([#12](https://github.com/terraform-providers/terraform-provider-ucloud/issues/12))
* **New Resource:** `ucloud_lb_ssl` ([#12](https://github.com/terraform-providers/terraform-provider-ucloud/issues/12))
* **New Resource:** `ucloud_lb_ssl_attachment` ([#12](https://github.com/terraform-providers/terraform-provider-ucloud/issues/12))
* **New Datasource:** `ucloud_instances` ([#12](https://github.com/terraform-providers/terraform-provider-ucloud/issues/12))
* **New Resource:** `ucloud_udpn_connection` ([#7](https://github.com/terraform-providers/terraform-provider-ucloud/issues/7))

ENHANCEMENTS:

* resource/ucloud_disk_attachment: Update schema version for disk attachment ID ([#12](https://github.com/terraform-providers/terraform-provider-ucloud/issues/12))
* resource/ucloud_vpc: Add update logic to `cidr_blocks` ([#9](https://github.com/terraform-providers/terraform-provider-ucloud/issues/9))
* provider: Support shared credential file and named profile ([#11](https://github.com/terraform-providers/terraform-provider-ucloud/issues/11))
* provider: Support customize endpoint url ([#11](https://github.com/terraform-providers/terraform-provider-ucloud/issues/11))

BUG FIXES:

* resource/ucloud_instance: Fix read of `image_id` and `instance_type` ([#12](https://github.com/terraform-providers/terraform-provider-ucloud/issues/12))
* resource/ucloud_instance: Check and create default firewall for new account ([#9](https://github.com/terraform-providers/terraform-provider-ucloud/issues/9))
* resource/ucloud_vpc: Fix cannot add multi value to `cidr_blocks` ([#9](https://github.com/terraform-providers/terraform-provider-ucloud/issues/9))

## 1.1.0 (January 09, 2019)

ENHANCEMENTS:

* resource/ucloud_eip_association: Update schema version for eip association `ID` ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_eip_association: Deprecated `resource_type` ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_lb_attachment: Deprecated `resource_type` ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_eip: Add `public_ip` attribute ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_instance: Update `instance_type` about customized ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* provider: Add `UserAgent` to external API ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))

BUG FIXES:

* resource/ucloud_disk: Fix default of `name` argument ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_eip: Fix default of `name` argument ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_instance: Fix default of `name` argument ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_lb_listener: Fix default of `name` argument ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_lb: Fix default of `name` argument ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_security_group: Fix default of `name` argument ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_subnet: Fix default of `name` argument ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))
* resource/ucloud_vpc: Fix default of `name` argument ([#2](https://github.com/terraform-providers/terraform-provider-ucloud/issues/2))

## 1.0.0 (December 19, 2018)

FEATURES:

* **New Resource:** `ucloud_instance`
* **New Resource:** `ucloud_disk`
* **New Resource:** `ucloud_disk_attachment`
* **New Resource:** `ucloud_eip`
* **New Resource:** `ucloud_eip_association`
* **New Resource:** `ucloud_security_group`
* **New Resource:** `ucloud_vpc`
* **New Resource:** `ucloud_subnet`
* **New Resource:** `ucloud_vpc_peering_connection`
* **New Resource:** `ucloud_lb`
* **New Resource:** `ucloud_lb_listener`
* **New Resource:** `ucloud_lb_attachment`
* **New Resource:** `ucloud_lb_rule`
* **New Datasource:** `ucloud_eips`
* **New Datasource:** `ucloud_images`
* **New Datasource:** `ucloud_projects`
* **New Datasource:** `ucloud_zones`
