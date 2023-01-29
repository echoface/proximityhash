# proximityhash
generates a set of geohashes that cover a circular area, given the center coordinates and the radius

a go version of proximityhash inspired by [python proximityhash](https://github.com/ashwin711/proximityhash)

addition a re-impliment Compression(GeoRaptor), it creates the best combination
of geohashes across various levels to represent a polygon, by starting from the
highest level and iterating till the optimal blend is brewed. Result accuracy
remains the same as that of the starting geohash level, but data size reduces
considerably for large polygons, thereby improving speed and performance.

Following is a sample of what georaptor does

![input](https://raw.github.com/ashwin711/georaptor/master/images/sgp_input.png)

![output](https://raw.github.com/ashwin711/georaptor/master/images/sgp_output.png)
