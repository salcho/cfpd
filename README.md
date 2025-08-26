# CCSDS FILE DELIVERY PROTOCOL

This repo holds a naive, partial implementation of the CCSDS File Delivery Protocol in Go. It implements a very small subset of the full spec, including:

- Unacknowledged transaction types.
- A minimal set of PDUs needed to support Copy File operations.
- TLV, LV value encoding.
- End-to-end testing for a `ListDirectory` operation.
- Support for UDP transport.
- CRC32, Proximity-1 CRC and modular checksums.

## Links

- [CFPD spec](https://ccsds.org/Pubs/727x0b5e1.pdf)
- Notes & papers on real-world deployments:
  - [James Webb Space Telescope - L2 Communications for Science Data Processing](https://ntrs.nasa.gov/api/citations/20080030196/downloads/20080030196.pdf)
  - [LRO (Lunar Reconnaissance Orbiter)](https://www.eoportal.org/satellite-missions/lro#sensor-complement)
  - [NASA Ground Network Support of the Lunar Reconnaissance Orbiter](https://web.archive.org/web/20120501145954/http://csse.usc.edu/gsaw/gsaw2007/s6/schupler.pdf)
  - [Ground-space communication solution for the European Space Agency (ESA)](https://arobs.com/blog/ground-space-communication-solution-for-the-european-space-agency-esa/#:~:text=Large%20amounts%20of%20data%20are,transfer%20protocol%20for%20future%20missions.)
  - [File Based Operations - Architectures and the EUCLID Example](https://arc.aiaa.org/doi/pdf/10.2514/6.2014-1750)
