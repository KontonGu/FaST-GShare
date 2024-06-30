
clean_crd_gen:
	rm -r pkg/client && rm -r pkg/apis/fastgshare.caps.in.tum/v1/zz_generated.deepcopy.go

gen_crd:
	bash code-gen.sh