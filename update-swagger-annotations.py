import os
import re

def update_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original_content = content
    
    # Replace imports
    content = re.sub(r'import io\.swagger\.annotations\.Api;', 'import io.swagger.v3.oas.annotations.tags.Tag;', content)
    content = re.sub(r'import io\.swagger\.annotations\.ApiOperation;', 'import io.swagger.v3.oas.annotations.Operation;', content)
    content = re.sub(r'import io\.swagger\.annotations\.ApiModel;', 'import io.swagger.v3.oas.annotations.media.Schema;', content)
    content = re.sub(r'import io\.swagger\.annotations\.ApiModelProperty;', 'import io.swagger.v3.oas.annotations.media.Schema;', content)
    content = re.sub(r'import io\.swagger\.annotations\.ApiImplicitParam;', 'import io.swagger.v3.oas.annotations.Parameter;', content)
    content = re.sub(r'import io\.swagger\.annotations\.\*;', 'import io.swagger.v3.oas.annotations.tags.Tag;\nimport io.swagger.v3.oas.annotations.Operation;\nimport io.swagger.v3.oas.annotations.media.Schema;\nimport io.swagger.v3.oas.annotations.Parameter;', content)
    
    # Replace @Api with @Tag
    content = re.sub(r'@Api\(tags\s*=\s*"([^"]+)"\)', r'@Tag(name = "\1")', content)
    content = re.sub(r'@Api\s*\(\s*tags\s*=\s*"([^"]+)"\s*\)', r'@Tag(name = "\1")', content)
    
    # Replace @ApiOperation with @Operation
    content = re.sub(r'@ApiOperation\("([^"]+)"\)', r'@Operation(summary = "\1")', content)
    content = re.sub(r'@ApiOperation\(value\s*=\s*"([^"]+)"\)', r'@Operation(summary = "\1")', content)
    
    # Replace @ApiModel with @Schema
    content = re.sub(r'@ApiModel\(description\s*=\s*"([^"]+)"\)', r'@Schema(description = "\1")', content)
    
    # Replace @ApiModelProperty with @Schema
    content = re.sub(r'@ApiModelProperty\(name\s*=\s*"([^"]+)",\s*value\s*=\s*"([^"]+)",\s*required\s*=\s*([^,]+),\s*dataType\s*=\s*"([^"]+)"\)', r'@Schema(name = "\1", description = "\2", requiredMode = \3 ? Schema.RequiredMode.REQUIRED : Schema.RequiredMode.NOT_REQUIRED, type = "\4")', content)
    content = re.sub(r'@ApiModelProperty\(name\s*=\s*"([^"]+)",\s*value\s*=\s*"([^"]+)",\s*dataType\s*=\s*"([^"]+)"\)', r'@Schema(name = "\1", description = "\2", type = "\3")', content)
    content = re.sub(r'@ApiModelProperty\(name\s*=\s*"([^"]+)",\s*value\s*=\s*"([^"]+)"\)', r'@Schema(name = "\1", description = "\2")', content)
    content = re.sub(r'@ApiModelProperty\(value\s*=\s*"([^"]+)",\s*required\s*=\s*([^,]+),\s*dataType\s*=\s*"([^"]+)"\)', r'@Schema(description = "\1", requiredMode = \2 ? Schema.RequiredMode.REQUIRED : Schema.RequiredMode.NOT_REQUIRED, type = "\3")', content)
    content = re.sub(r'@ApiModelProperty\(value\s*=\s*"([^"]+)",\s*dataType\s*=\s*"([^"]+)"\)', r'@Schema(description = "\1", type = "\2")', content)
    content = re.sub(r'@ApiModelProperty\(value\s*=\s*"([^"]+)"\)', r'@Schema(description = "\1")', content)
    content = re.sub(r'@ApiModelProperty\(name\s*=\s*"([^"]+)",\s*dataType\s*=\s*"([^"]+)"\)', r'@Schema(name = "\1", type = "\2")', content)
    content = re.sub(r'@ApiModelProperty\(name\s*=\s*"([^"]+)"\)', r'@Schema(name = "\1")', content)
    
    # Replace @ApiImplicitParam with @Parameter
    content = re.sub(r'@ApiImplicitParam\(name\s*=\s*"([^"]+)",\s*value\s*=\s*"([^"]+)",\s*required\s*=\s*([^,]+),\s*dataType\s*=\s*"([^"]+)"\)', r'@Parameter(name = "\1", description = "\2", required = \3, schema = @Schema(type = "\4"))', content)
    content = re.sub(r'@ApiImplicitParam\(name\s*=\s*"([^"]+)",\s*value\s*=\s*"([^"]+)",\s*required\s*=\s*([^,]+)\)', r'@Parameter(name = "\1", description = "\2", required = \3)', content)
    content = re.sub(r'@ApiImplicitParam\(name\s*=\s*"([^"]+)",\s*value\s*=\s*"([^"]+)"\)', r'@Parameter(name = "\1", description = "\2")', content)
    
    # Handle special cases for requiredMode
    if 'Schema.RequiredMode.REQUIRED' in content or 'Schema.RequiredMode.NOT_REQUIRED' in content:
        if 'import io.swagger.v3.oas.annotations.media.Schema;' not in content:
            content = 'import io.swagger.v3.oas.annotations.media.Schema;\n' + content
    
    if content != original_content:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f'Updated: {file_path}')
        return True
    return False

def main():
    src_dir = r'c:\Users\aqi\Desktop\aurora-master\aurora-springboot\src\main\java'
    count = 0
    
    for root, dirs, files in os.walk(src_dir):
        for file in files:
            if file.endswith('.java'):
                file_path = os.path.join(root, file)
                if update_file(file_path):
                    count += 1
    
    print(f'\nTotal files updated: {count}')

if __name__ == '__main__':
    main()
