include scripts/make-rules/common.mk # make sure include common.mk at the first include line

# select the cmd which contains 'apiserver'
#API_COMMANDS ?= $(filter %apiserver, $(wildcard ${ROOT_DIR}/cmd/*))
#API_BINS ?= $(foreach cmd,${API_COMMANDS},$(notdir ${cmd}))

CONFIGS_DIR := $(ROOT_DIR)/configs
#LOCAL_CONFIG := $(addsuffix .yaml,${CONFIGS_DIR}/local-${API_BINS})

#.PHONY: local.run
#local.run:
#	$(eval TARGET_NAME := "apiserver")
#    $(eval LOCAL_CONFIG := $(addsuffix .yaml,${CONFIGS_DIR}/local-${TARGET_NAME}))
#    $(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
#    $(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
#
#	@make gen
#	@make go.build.$(PLATFORM).$(API_BINS)
#	$(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)/$(API_BINS) -c $(LOCAL_CONFIG)


# 使用模式规则（Pattern Rules）支持 make local.run.% 的语法
.PHONY: local.run.%
local.run.%:
	$(eval TARGET_NAME := $*)
	$(eval LOCAL_CONFIG := $(addsuffix .yaml,${CONFIGS_DIR}/local-${TARGET_NAME}))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))

	@echo "Building and running ${TARGET_NAME} with config ${LOCAL_CONFIG}"
	@make gen
	@make go.build.$(PLATFORM).$(TARGET_NAME)
	@env $(shell cat $(ROOT_DIR)/env/local | xargs) $(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)/$(TARGET_NAME) -c $(LOCAL_CONFIG)
