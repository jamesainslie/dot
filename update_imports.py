#!/usr/bin/env python3
"""Update imports from pkg/dot to internal/domain in internal packages."""

import os
import re
from pathlib import Path

def update_imports_in_file(filepath):
    """Update imports and qualified identifiers in a single file."""
    with open(filepath, 'r') as f:
        content = f.read()
    
    original = content
    
    # Replace the import statement
    content = content.replace(
        '"github.com/jamesainslie/dot/pkg/dot"',
        '"github.com/jamesainslie/dot/internal/domain"'
    )
    
    # Replace qualified identifiers dot. with domain.
    # Only replace if the import was pkg/dot (check if domain import exists)
    if 'github.com/jamesainslie/dot/internal/domain' in content:
        # Replace dot.Something with domain.Something
        # Use word boundary to avoid replacing things like "dot.go" in comments
        content = re.sub(r'\bdot\.', 'domain.', content)
    
    # If content changed, write it back
    if content != original:
        with open(filepath, 'w') as f:
            f.write(content)
        return True
    return False

def main():
    """Update all files in internal/ packages."""
    root = Path('/Volumes/Development/dot')
    internal_dir = root / 'internal'
    
    # Packages to update (exclude internal/domain and internal/api)
    packages = [
        'executor',
        'pipeline', 
        'scanner',
        'planner',
        'manifest',
        'config',
        'adapters',
        'ignore',
        'cli'
    ]
    
    updated_files = []
    
    for package in packages:
        package_dir = internal_dir / package
        if not package_dir.exists():
            continue
            
        # Find all .go files
        for gofile in package_dir.rglob('*.go'):
            if update_imports_in_file(gofile):
                updated_files.append(str(gofile.relative_to(root)))
    
    print(f"Updated {len(updated_files)} files:")
    for f in updated_files:
        print(f"  - {f}")

if __name__ == '__main__':
    main()