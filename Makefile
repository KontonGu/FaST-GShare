
clean_crd_gen:
	rm -r pkg/client && rm -r pkg/apis/fastgshare.caps.in.tum/v1/zz_generated.deepcopy.go

gen_crd:
	bash code-gen.sh

code_gen_crd:
	cd .. && git clone https://github.com/kubernetes/code-generator.git && cd code-generator && git checkout release-1.23
	bash code-gen.sh