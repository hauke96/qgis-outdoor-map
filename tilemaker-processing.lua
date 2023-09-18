-- Nodes will only be processed if one of these keys is present
node_keys = { "amenity", "shop" }

-- Initialize Lua logic (optional)
function init_function()
end

-- Finalize Lua logic (optional)
function exit_function()
end

-- Assign nodes to a layer, and set attributes, based on OSM tags
function node_function(node)
	local amenity = node:Find("amenity")
	local shop = node:Find("shop")
	if amenity~="" or shop~="" then
		node:Layer("poi", false)
		-- node:Attribute("name", node:Find("name"))
		-- if amenity~="" then node:Attribute("class",amenity)
		-- else node:Attribute("class",shop) end
	end
end

-- Assign ways to a layer, and set attributes, based on OSM tags
function way_function(way)
	local highway = way:Find("highway")
	if highway~="" then
		way:Layer("highway", false)
		add_tag(way, "highway", "name", "smoothness", "trail_visibility", "sac:scale", "surface", "tracktype")
		return
	end

	local landuse = way:Find("landuse")
	if landuse~="" then
		way:Layer("landuse", true)
		add_tag(way, "landuse", "name")
		return
	end

	local natural = way:Find("natural")
	if natural~="" then
		-- Put "natural" things also to landuse for convenience
		way:Layer("landuse", true)
		add_tag(way, "name")
		way:Attribute("landuse", natural)
		return
	end

	local building = way:Find("building")
	if building~="" then
		way:Layer("building", true)
		return
	end
end

function add_tag(feature, ...)
	local tags = {...}
	for i = 1, #tags do
		feature:Attribute(tags[i], feature:Find(tags[i]))
	end
end
