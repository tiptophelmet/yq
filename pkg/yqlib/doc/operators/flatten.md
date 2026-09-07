# Flatten
This recursively flattens arrays.

## Flatten
Recursively flattens all arrays

Given a sample.yml file of:
```yaml
- 1
- - 2
- - - 3
```
then
```bash
yq 'flatten' sample.yml
```
will output
```yaml
- 1
- 2
- 3
```

## Flatten with depth of one
Given a sample.yml file of:
```yaml
- 1
- - 2
- - - 3
```
then
```bash
yq 'flatten(1)' sample.yml
```
will output
```yaml
- 1
- 2
- - 3
```

## Flatten empty array
Given a sample.yml file of:
```yaml
- []
```
then
```bash
yq 'flatten' sample.yml
```
will output
```yaml
[]
```

## Flatten array of objects
Given a sample.yml file of:
```yaml
- foo: bar
- - foo: baz
```
then
```bash
yq 'flatten' sample.yml
```
will output
```yaml
- foo: bar
- foo: baz
```

## Flatten resolves an alias to a sequence
Aliased sequences are dereferenced and inlined, just like an inline sequence.

Given a sample.yml file of:
```yaml
base_list: &base
  - item1
  - item2
extended_list:
  - *base
  - item3
```
then
```bash
yq '.extended_list | flatten' sample.yml
```
will output
```yaml
- item1
- item2
- item3
```

## Flatten leaves aliases to non-sequences untouched
Only aliases that resolve to sequences are dereferenced; other aliases are left as-is.

Given a sample.yml file of:
```yaml
m: &m
  a: 1
s: &s cat
l:
  - *m
  - *s
  - 2
```
then
```bash
yq '.l | flatten' sample.yml
```
will output
```yaml
- *m
- *s
- 2
```

