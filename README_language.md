
<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubectl-sql/master/img/kubesql-162.png" alt="kubectl-sql Logo">
</p>

# kubectl-sql

## Query language

kubectl-sql uses Tree Search Language (TSL). TSL is a wonderful human readable filtering language.

https://github.com/yaacov/tree-search-language


#### Available Operators:

| Operators | Example |
|----|---|
| `=`, `~=` | `name ~= '^test-'`  |
| `like`, `ilike` | `phase ilike 'run%'`  | 
|`!=`, `!~` |  `namespace != 'default'` |
|`>`, `<`, `<=` and `>=` | `created < 2020-01-15T00:00:00Z` |
|`is null`, `is not null`| `spec.domain.cpu.dedicatedCpuPlacement is not null` |
| `in`   |  `spec.domain.resources.limits.memory in (1Gi, 2Gi)` |
| `between`   |  `spec.domain.resources.limits.memory between 1Gi and 2Gi` |
| `or`, `and` and `not` | `name ~= 'virt-' and namespace != 'test-wegsb'` |
| `( )`|  `phase = 'Running' and (namespace ~= 'cnv-' or namespace ~= 'virt-')`|

#### Available Math Operators:

| Operators | Notes |
|----|---|
| `+`, `-` | Addition and Subtraction |
| `*`, `/` | Multiplication and Division |
| `( )`|  |

#### Aliases:
| Alias | Resource Path | Example |
|----|---|---|
| name | metadata.name | |
| namespace | metadata.namespace | `namespace ~= '^test-[a-z]+$'` |
| labels | metadata.labels | |
| annotations | metadata.annotations | |
| created | creation timestamp | |
| deleted | deletion timestamp | |
| phase | status.phase | `phase = 'Running'` |

#### SI Units:
| Unit | Multiplier | Example |
|----|---|---|
| Ki | 1024 | |
| Mi | 1024^2 | `spec.containers[1].resources.requests.memory = 200Mi` |
| Gi | 1024^3 | |
| Ti | 1024^4 | |
| Pi | 1024^5 | |

#### Booleans:
| Example |
|---|
| `status.conditions[1].status = true` |

#### Dates:
| Format | Example |
|---|---|
| RFC3339 | `status.conditions[1].lastTransitionTime > 2020-02-20T11:12:38Z`  |
| Short date | `created <= 2020-02-20` |

#### Arrays and lists:
kubectl-sql support resource paths that include lists by using the list index as a field key.

``` bash
# Get the memory request for the first container.
kubectl-sql --all-namespaces "select * from pods where spec.containers[1].resources.requests.memory = 200Mi"
```
