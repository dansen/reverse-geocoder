# reverse-geocoder-go

Go implementation (work in progress) of the Python reverse_geocoder library.

## Features (Planned)
- Offline reverse geocoding using GeoNames cities1000 dataset
- KD-Tree based nearest city lookup
- Single-threaded and multi-threaded query modes
- Automatic data download & processing
- Custom data source support

## Project Structure
See `design.md` for the detailed architecture and file-level design.

## Quick Start (Coming Soon)
```bash
# build CLI
cd go/cmd/rgeocoder
go build -o rgeocoder
./rgeocoder -mode 2 37.78674 -122.39222
```

## License
Same as upstream project (MIT).