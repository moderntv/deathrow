[![Go Report Card](https://goreportcard.com/badge/github.com/moderntv/deathrow)](https://goreportcard.com/report/github.com/moderntv/deathrow)
![Go Version](https://img.shields.io/github/go-mod/go-version/moderntv/deathrow)
[![codecov](https://codecov.io/gh/moderntv/deathrow/branch/master/graph/badge.svg?token=Z2KMHW5AOR)](https://codecov.io/gh/moderntv/deathrow)

# DeathRow

DeathRow is a time-based priority queue, which only pops items from `Prison` after they reached their deadline.

## Why?

Slightly more efficient checking of TTLs in caches or work queues
