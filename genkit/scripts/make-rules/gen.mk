# Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file.

# ==============================================================================
# Makefile helper functions for generate necessary files
#

GO := go

.PHONY: gen.run
#gen.run: gen.errcode gen.docgo
gen.run: gen.clean gen.errcode gen.docgo.doc

.PHONY: gen.errcode
gen.errcode: gen.errcode.code

.PHONY: gen.errcode.code
gen.errcode.code: tools.verify.codegen
	@echo "===========> Generating error code files for listingagent"
	goerr-gen -output=${ROOT_DIR}/apps/listingagent/model/code \
		-docOutput=${ROOT_DIR}/docs/api/error_code_generated.md \
		${ROOT_DIR}/apps/listingagent/model/code ${ROOT_DIR}/apps/pkg/code

.PHONY: gen.doc.code
gen.doc.code: gen.errcode.code

.PHONY: gen.docgo.doc
gen.docgo.doc:
	@echo "===========> Generating missing doc.go for go packages"
	@${ROOT_DIR}/scripts/gendoc.sh

.PHONY: gen.docgo.check
gen.docgo.check: gen.docgo.doc
	@n="$$(git ls-files --others '*/doc.go' | wc -l)"; \
	if test "$$n" -gt 0; then \
		git ls-files --others '*/doc.go' | sed -e 's/^/  /'; \
		echo "$@: untracked doc.go file(s) exist in working directory" >&2 ; \
		false ; \
	fi

.PHONY: gen.docgo.add
gen.docgo.add:
	@git ls-files --others '*/doc.go' | $(XARGS) -- git add

.PHONY: gen.defaultconfigs
gen.defaultconfigs:
	@${ROOT_DIR}/scripts/gen_default_config.sh

.PHONY: gen.doc.cmd.%
gen.doc.cmd.%:
	$(eval APP := $*)
	@echo "===========> Generating docs for $(APP) cmd"
	@$(GO) run $(ROOT_DIR)/cmd/$(APP)/$(APP).go -d $(ROOT_DIR)/docs/cmd
	@echo "<=========== Please check docs at docs/cmd"

.PHONY: gen.doc.cmd
gen.doc.cmd: $(addprefix gen.doc.cmd., $(foreach app,${APPS},$(notdir ${app})))

.PHONY: gen.doc.api
gen.doc.api: swagger.run

.PHONY: gen.clean
gen.clean:
	@rm -rf ./api/client/{clientset,informers,listers}
	@$(FIND) -type f -name '*_generated.go' -delete


.PHONY: gen.db.%
gen.db.%:
	$(eval APP := $*)
	@echo "===========> Generating db files for apps/${APP}"
	@GO run tools/gendb/gendb.go ${ROOT_DIR}/apps/${APP}/internal/domain
