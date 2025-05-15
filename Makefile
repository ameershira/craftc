# Makefile with dependency tracking

# Compiler and flags
CC := cc
CFLAGS := -MMD -MP

# Directory paths
SRC_DIR := ./libgensrc
OBJ_DIR := ./build/maketest1/obj

# Find all .c source files
SRCS := $(wildcard $(SRC_DIR)/*.c)

# Compute object and dependency files
OBJS := $(patsubst $(SRC_DIR)/%.c, $(OBJ_DIR)/%.o, $(SRCS))
DEPS := $(OBJS:.o=.d)

# Default target: compile all object files
all: $(OBJS)

# Rule to compile .c to .o and generate .d files
$(OBJ_DIR)/%.o: $(SRC_DIR)/%.c
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -c $< -o $@

# Clean rule
clean:
	rm -rf $(OBJ_DIR)

# Include generated dependency files if they exist
-include $(DEPS)

.PHONY: all clean
