
<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubectl-sql/master/img/kubesql-162.png" alt="kubectl-sql Logo">
</p>

# kubectl-sql

## Config File

<p align="center">
  <a href="https://asciinema.org/a/308440" target="_blank"><img src="https://asciinema.org/a/308440.svg" /></a>
</p>

Users can add aliases and edit the fields displayed in table view using json config files,
[see the example config file](https://github.com/yaacov/kubectl-sql/blob/master/kubectl-sql.json).

Flag: `--kubectl-sql <config file path>` (default: `$HOME/.kube/kubectl-sql.json`)

Example:

``` bash
kubectl-sql --kubectl-sql ./kubectl-sql.json get pods
...
```
