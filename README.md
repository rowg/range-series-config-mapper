# HF Radar Config Mapper

## Getting started
### Setup conda environment
```
mamba env create -f environment.yml
mamba activate hfradar-config-mapper
```

### Building the binary
In the project root directory, run the following command:
```
go build
```

A binary will appear in the project directory: `hfradar-config-mapper`

To compile for a specific os/platform, set the `GOOS` & `GOARCH` environment variables:
```
GOOS=linux GOARCH=amd64 go build
```

## Usage
### HFR Site Directory Structure
The HF Radar archive data is assumed to be structured in the following manner:
```
- Operator/
  - Site/
    - RangeSeries/
    - Config_Auto/
    - Config_Operator/
```

### Flags
`hfradar-config-mapper` accepts the following CLI flags:
- `--site-dir`: The directory of the HF Radar site that you want to create a RangeSeries:Config mapping for
- `--output-file-type`: The desired file format for the output, either `JSON` or `CSV`
- `--output-file-name`: The base name for the output file.
- `-all`: Boolean flag indicating whether to produce a mapping for all RangeSeries files for the site. If set, `siteDir/RangeSeries` will be scanned for RangeSeries files.

### Arguments
You can specify the RangeSeries files of interest by passing them as unnamed arguments after the flags. When the `-all` flag is not set, a mapping will be created for the RangeSeries files that are passed in this manner.

### Examples
Compute mapping for all config files. Output result in `JSON` format to `myMapping.json`:
```
./hfradar-config-mapper \
    --site-dir="/my/hfradar/archive/dir/UCSB/MGS1" \
    --output-file-type="JSON" \
    --output-file-name="myMapping" \
    -all
```

Compute mapping for `Rng_mgs1_2023_05_17_070610.rs`. Output result in `CSV` format as `mgs1_configs.csv`.
```
./hfradar-config-mapper \
    --site-dir="/my/hfradar/archive/dir/UCSB/MGS1" \
    --output-file-type="CSV" \
    --output-file-name="mgs1_configs" \
    /my/hfradar/archive/dir/UCSB/MGS1/RangeSeries/2023/05/17/Rng_mgs1_2023_05_17_070610.rs
```

Compute mapping for both `Rng_mgs1_2023_05_17_070610.rs` and `Rng_mgs1_2023_05_23_032006.rs`. Output result in `JSON` format as `mgs1_configs.json`.
```
./hfradar-config-mapper \
    --site-dir="/my/hfradar/archive/dir/UCSB/MGS1" \
    --output-file-type="JSON" \
    --output-file-name="mgs1_configs" \
    /my/hfradar/archive/dir/UCSB/MGS1/RangeSeries/2023/05/17/Rng_mgs1_2023_05_17_070610.rs \
    /my/hfradar/archive/dir/UCSB/MGS1/RangeSeries/2023/05/23/Rng_mgs1_2023_05_23_032006.rs
```