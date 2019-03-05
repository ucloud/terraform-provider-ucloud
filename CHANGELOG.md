## 1.2.0 (Unreleased)

FEATURES:

* **New Resource:** `ucloud_db_instance` [GH-12]
* **New Resource:** `ucloud_lb_ssl` [GH-12]
* **New Resource:** `ucloud_lb_ssl_attachment` [GH-12]
* **New Datasource:** `ucloud_instances` [GH-12]
* **New Resource:** `ucloud_udpn_connection` [GH-7]

ENHANCEMENTS:

* resource/ucloud_disk_attachment: Update schema version for disk attachment ID [GH-12]
* resource/ucloud_vpc: Add update logic to `cidr_blocks` [GH-9]
* provider: Support shared credential file and named profile [GH-11]
* provider: Support customize endpoint url [GH-11]

BUG FIXES:

* resource/ucloud_instance: Fix read of `image_id` and `instance_type` [GH-12]
* resource/ucloud_instance: Check and create default firewall for new account [GH-9]
* resource/ucloud_vpc: Fix cannot add multi value to `cidr_blocks` [GH-9]

## 1.1.0 (January 09, 2019)

ENHANCEMENTS:

* resource/ucloud_eip_association: Update schema version for eip association `ID` ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_eip_association: Deprecated `resource_type` ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_lb_attachment: Deprecated `resource_type` ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_eip: Add `public_ip` attribute ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_instance: Update `instance_type` about customized ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* provider: Add `UserAgent` to external API ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))

BUG FIXES:

* resource/ucloud_disk: Fix default of `name` argument ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_eip: Fix default of `name` argument ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_instance: Fix default of `name` argument ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_lb_listener: Fix default of `name` argument ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_lb: Fix default of `name` argument ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_security_group: Fix default of `name` argument ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_subnet: Fix default of `name` argument ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))
* resource/ucloud_vpc: Fix default of `name` argument ([#2](https://github.com/ucloud/terraform-provider-ucloud/issues/2))

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
