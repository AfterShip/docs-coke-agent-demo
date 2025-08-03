# Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file.

# ==============================================================================
# Makefile helper functions for swagger
#

include scripts/make-rules/common.mk

APP_NAMES := $(patsubst $(ROOT_DIR)/apps/%, %, $(APPS))

.PHONY: swagger.run
swagger.run:
	@for app in $(APP_NAMES); do \
		$(MAKE) swagger.run.$$app; \
	done

swagger.run.%:
	$(eval app := $(*))
	@echo "===========> Generating Swagger for app: $(app)";
	@swagger generate spec --scan-models -w  $(ROOT_DIR)/apps/$(app) -o $(ROOT_DIR)/docs/api/$(app)_swagger.json;
	@echo -e "<=========== Please check docs at $(ROOT_DIR)/docs/api/$(app)_swagger.json\n"


.PHONY: swagger.serve
swagger.serve:
	@swagger serve -F=redoc --no-open --port 36666 $(ROOT_DIR)/docs/api/apiserver_swagger.json
