# ./code-generator/generate-groups.sh all \
# "./pkg/client" "./pkg/apis" --go-header-file="./hack/boilerplate.go.txt" fastgshare.caps.in.tum:v1 --output-base=./

../code-generator/generate-groups.sh all github.com/KontonGu/FaST-GShare/pkg/client \
github.com/KontonGu/FaST-GShare/pkg/apis fastgshare.caps.in.tum.de:v1 --go-header-file=boilerplate.go.txt --output-base=../../../


