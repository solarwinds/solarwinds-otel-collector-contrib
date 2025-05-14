# OS details attributes

Attributes describing more detailed information about operating system. The OS Details consist of InfoStat and Language attributes.
## Attribute details

| Name                            | Description                       | Values  | Example                               |
| ------------------------------- | --------------------------------- | ------- | ------------------------------------- |
| osdetails.hostname              | Host name                         | Any Str | ...                                   |
| osdetails.boottime              | Host boottime in unix format      | Any Int | 1677757113                            |
| osdetails.os                    | Operating system                  | Any Str | windows                               |
| osdetails.platform              | Operating system platform         | Any Str | Microsoft Windows 10 Enterprise       |
| osdetails.platform.family       | Operating system platform family  | Any Str | Standalone Workstation                |
| osdetails.platform.version      | Operation system platform version | Any Str | 10.0.19045 Build 19045                |
| osdetails.kernel.version        | Kernel version                    | Any Str | 10.0.19045 Build 19045                |
| osdetails.kernel.architecture   | Kernel architecture               | Any Str | x86_64                                |
| osdetails.virtualization.system | Virtualization system             | Any Str | xen                                   |
| osdetails.virtualization.role   | Virtualization role               | Any Str | guest                                 |
| osdetails.host.id               | Host ID                           | Any Str | 482f3116-df31-4b05-b74c-4207a2b8500b  |

## Windows specific attributes
For Windows we poll attributes listed above. Along these we poll some Windows specific attributes.

| Name                            | Description                       | Values  | Example                               |
| ------------------------------- | --------------------------------- | ------- | ------------------------------------- |
| osdetails.language.lcid         | Locale ID                         | Any Int | 1033                                  |
| osdetails.language.name         | Language Name                     | Any Str | en-US                                 |
| osdetails.language.displayname  | Language display name             | Any Str | English (United States)               |

## Linux specific attributes
For Linux we poll attributes listed above. Along these we poll some Linux specific attributes.

| Name                            | Description                       | Values  | Example                               |
| ------------------------------- | --------------------------------- | ------- | ------------------------------------- |
| osdetails.language.name         | LANGUAGE environment variable     | Any Str | en_US                                 |
