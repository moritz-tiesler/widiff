set -euo pipefail

export $(xargs -a ../.env)

go run main.go

unset $(xargs -a ../.env | sed 's/=.*//')

exit
