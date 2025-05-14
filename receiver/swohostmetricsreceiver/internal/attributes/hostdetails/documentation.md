# OS details attributes

Attributes describing more detailed information about host. The Host Details consist of Domain, Model and Time Zone attributes.
## Windows attribute details

| Name                              | Description                                             | Values  | Example                                                       |
|-----------------------------------|---------------------------------------------------------| ------- |---------------------------------------------------------------|
| hostdetails.domain                | Host domain                                             | Any Str | swdev.local                                                   |
| hostdetails.domain.fqdn           | Fully Qualified Domain Name                             | Any Str | SWI-D5ZRKQ2.swdev.local                                       |
| hostdetails.domain.role           | Domain role                                             | Any Int | 1                                                             |
| hostdetails.model.serialnumber    | Model serial number                                     | Any Str | D5ZRKQ2                                                       |
| hostdetails.model.manufacturer    | Model manufacturer                                      | Any Str | Dell Inc.                                                     |
| hostdetails.model.name            | Model name                                              | Any Str | OptiPlex 7060                                                 |
| hostdetails.timezone.bias         | Difference in minutes of between the local time and UTC | Any Int | 120                                                           |
| hostdetails.timezone.caption      | Caption of the local time zone                          | Any Str | (UTC+01:00) Belgrade, Bratislava, Budapest, Ljubljana, Prague |
| hostdetails.timezone.standardname | Standard name of the local time zone                    | Any Str | Central Europe Standard Time                                  |

## Linux attribute details

| Name                              | Description                                             | Values  | Example              |
|-----------------------------------|---------------------------------------------------------| ------- |----------------------|
| hostdetails.domain                | Host domain                                             | Any Str | ec2.internal         |
| hostdetails.domain.fqdn           | Fully Qualified Domain Name                             | Any Str | SWI-D5ZRKQ2          |
| hostdetails.timezone.caption      | Caption of the local time zone                          | Any Str | Etc/UTC (UTC, +0000) |
