<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubectl-sql/master/img/kubesql-162.png" alt="kubectl-sql Logo">
</p>

# kubectl‑sql — Query Language Reference

kubectl‑sql uses **Tree Search Language (TSL)** – a human‑readable filtering grammar shared with the [`tree-search-language`](https://github.com/yaacov/tree-search-language) project.  The tables below document operators, literals and helper supported by the TSL.

---

## Operators

| Category         | Operators                                                         | Example                                                         |
| ---------------- | ----------------------------------------------------------------- | --------------------------------------------------------------- |
| Equality / regex | `=`, `!=`, `~=` *(regex match)*, `~!` *(regex ****not**** match)* | `name ~! '^test-'`                                              |
| Pattern          | `like`, `not like`, `ilike`, `not ilike`                          | `phase not ilike 'run%'`                                        |
| Comparison       | `>`, `<`, `>=`, `<=`                                              | `created < 2020‑01‑15T00:00:00Z`                                |
| Null tests       | `is null`, `is not null`                                          | `spec.domain.cpu.dedicatedCpuPlacement is not null`             |
| Membership       | `in`, `not in`                                                    | `memory in [1Gi, 2Gi]`                                          |
| Ranges           | `between`, `not between`                                          | `memory between 1Gi and 4Gi`                                    |
| Boolean          | `and`, `or`, `not`                                                | `name ~= 'virt-' and not namespace = 'default'`                 |
| Grouping         | `( … )`                                                           | `(phase='Running' or phase='Succeeded') and namespace~='^cnv-'` |

---

## Math & Unary Operators

| Operator      | Description                                                   |
| ------------- | ------------------------------------------------------------- |
| `+`, `-`      | Addition & subtraction *(prefix **`+x`** / **`-x`** allowed)* |
| `*`, `/`, `%` | Multiplication, division, modulo                              |
| `( … )`       | Parentheses to override precedence                            |

---

## Aliases

| Alias         | Resource path          | Example                      |
| ------------- | ---------------------- | ---------------------------- |
| `name`        | `metadata.name`        | `name ~= '^test-'`           |
| `namespace`   | `metadata.namespace`   | `namespace != 'kube-system'` |
| `labels`      | `metadata.labels`      | `labels.env = 'prod'`        |
| `annotations` | `metadata.annotations` |                              |
| `created`     | creationTimestamp      | `created > 2023‑01‑01`       |
| `deleted`     | deletionTimestamp      |                              |
| `phase`       | `status.phase`         | `phase = 'Running'`          |

---

## Size & Time Literals

### SI / IEC units

#### SI units (powers of 1000)

| Suffix | Multiplier |
| ------ | ---------- |
| k / K  | 10³         |
| M      | 10⁶         |
| G      | 10⁹         |
| T      | 10¹²        |
| P      | 10¹⁵        |

#### IEC units (powers of 1024)

| Suffix | Multiplier |
| ------ | ---------- |
| Ki     | 1024¹       |
| Mi     | 1024²       |
| Gi     | 1024³       |
| Ti     | 1024⁴       |
| Pi     | 1024⁵       |

### Scientific notation

Numbers may be written as `6.02e23`, `2.5E‑3`, etc.

---

## Booleans

The literals `true` and `false` (case‑insensitive) evaluate to boolean values.

---

## Dates

| Format     | Example                                     |
| ---------- | ------------------------------------------- |
| RFC 3339   | `lastTransitionTime > 2025‑02‑20T11:12:38Z` |
| Short date | `created <= 2025‑02‑20`                     |

---

## Arrays & Lists

Fields may include list indices, wildcards or named keys:

```tsl
spec.containers[0].resources.requests.memory = 200Mi
spec.ports[*].protocol = 'TCP'
spec.ports[http‑port].port = 80
```

### Membership tests with lists

Use **square‑bracket literals** when testing membership:

```tsl
memory in [1Gi, 2Gi, 4Gi]
```

### Array helpers

| Helper | Example                                                     |
| ------ | ----------------------------------------------------------- |
| `any`  | `any (spec.containers[*].resources.requests.memory = 200Mi)` |
| `all`  | `all (spec.containers[*].resources.requests.memory != null)`  |
| `len`  | `len spec.containers[*] > 2`                               |

`any`, `all`, and `len` may be called *with or without* parentheses: `any expr` is equivalent to `any(expr)`.

---

> **Tip – mixing selectors**: Combine aliases, regex, math and list helpers to build expressive filters, e.g.
>
> ```tsl
> any(phase = 'Running') and namespace ~= '^(cnv|virt)-'
> ```
