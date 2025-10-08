#!/usr/bin/env python3
"""Fix renderer files to import both domain and dot."""

from pathlib import Path
import re

files = [
    'internal/cli/renderer/json.go',
    'internal/cli/renderer/table.go',
    'internal/cli/renderer/text.go',
    'internal/cli/renderer/yaml.go',
    'internal/cli/renderer/text_test.go',
    'internal/cli/renderer/renderer_test.go',
]

root = Path('/Volumes/Development/dot')

for filepath in files:
    full_path = root / filepath
    if not full_path.exists():
        continue
        
    with open(full_path, 'r') as f:
        content = f.read()
    
    # Add pkg/dot import if not present and domain import exists
    if 'internal/domain' in content and 'pkg/dot' not in content:
        # Find the import block and add pkg/dot
        content = content.replace(
            '"github.com/jamesainslie/dot/internal/domain"',
            '"github.com/jamesainslie/dot/internal/domain"\n\t"github.com/jamesainslie/dot/pkg/dot"'
        )
    
    # Replace domain.Status with dot.Status
    content = content.replace('domain.Status', 'dot.Status')
    # Replace domain.PackageInfo with dot.PackageInfo
    content = content.replace('domain.PackageInfo', 'dot.PackageInfo')
    
    with open(full_path, 'w') as f:
        f.write(content)

print("Fixed renderer files")
