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
	if amenity ~= "" or shop ~= "" then
		node:Layer("poi", false)
		-- node:Attribute("name", node:Find("name"))
		-- if amenity ~= "" then node:Attribute("class",amenity)
		-- else node:Attribute("class",shop) end
	end
end

-- Assign ways to a layer, and set attributes, based on OSM tags
function way_function(way)
	local highway = way:Find("highway")
	if highway ~= "" then
		way:Layer("highway", false)
		add_tag(way, "highway", "name", "smoothness", "trail_visibility", "sac:scale", "via_ferrata_scale", "surface", "tracktype")
		return
	end

	local landuse = way:Find("landuse")
	if landuse ~= "" then
		way:Layer("landuse", true)
		add_tag(way, "landuse", "name")
		return
	end

	local natural = way:Find("natural")
	if natural ~= "" then
		-- Put "natural" things also to landuse for convenience
		way:Layer("landuse", true)
		add_tag(way, "name")
		way:Attribute("landuse", natural)
		return
	end

	local building = way:Find("building")
	if building ~= "" then
		way:Layer("building", true)
		return
	end

	local boundary = way:Find("boundary")
	if boundary ~= "" then
		way:Layer("boundary", false)
		add_tag(way, "boundary", "name", "protect_class", "border_type", "admin_level")
		return
	end

	local waterway = way:Find("waterway")
	if waterway ~= "" then
		way:Layer("waterway", false)
		add_tag(way, "waterway", "name", "intermittent", "tunnel")
		return
	end

	local railway = way:Find("railway")
	--if railway == "rail" or railway == "narrow_gauge" or railway == "funicular" or railway == "light_rail" or railway == "monorail" or railway == "subway" or railway == "tram" or railway == "disused" or railway == "abandoned" then
	if railway ~= "" then
		if any_of(railway, "rail", "narrow_gauge", "funicular", "light_rail", "monorail", "subway", "tram", "disused", "abandoned") then
			-- Consider these tags as lines
			way:Layer("railway", false)
			add_tag(way, "railway", "name", "tunnel")
			return
		else
			-- Everything else is considered to be a polygon, especially stations and platforms
			way:Layer("railway", true)
			add_tag(way, "railway", "name", "tunnel")
			return
		end
	end
end

function relation_function(relation)
	local boundary = relation:Find("boundary")
	if boundary ~= "" then
		relation:Layer("boundary", false)
		add_tag(relation, "boundary", "name", "protect_class", "border_type", "admin_level")
		return
	end
end

function add_tag(feature, ...)
	local tags = {...}
	for i = 1, #tags do
		feature:Attribute(tags[i], feature:Find(tags[i]))
	end
end

function any_of(value, ...)
	local others = {...}
	for i = 1, #others do
		if value == others[i]then
			return true
		end
	end
	return false
end

function relation_scan_function(relation)
	if relation:Find("type") == "boundary" then
		relation:Accept()
	end
end
