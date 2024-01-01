-- Nodes will only be processed if one of these keys is present
node_keys = {
	"amenity",
	"natural",
	"railway",
	"shop",
	"label",
	"place"
}

-- Route attributes to copy
route_attributes = { "name", "colour", "ref" }

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
	local natural = node:Find("natural")
	if amenity ~= "" or shop ~= "" or natural ~= "" then
		node:Layer("poi", false)
		add_tag(node, "name", "amenity", "shop", "natural", "ele")
	end

	local label = node:Find("label")
	local category = node:Find("category")
	if label == "yes" then
		node:Layer("label", false)
		add_tag(node, "type", "category", "text")
		return
	end

	local place = node:Find("place")
	if place ~= "" then
		node:Layer("label", false)
		add_tag(node, "place", "name")
		return
	end

	local railway = node:Find("railway")
	if railway ~= "" then
		node:Layer("railway", false)
		add_tag(node, "railway", "name")
		return
	end
end

-- Assign ways to a layer, and set attributes, based on OSM tags
function way_function(way)
	local highway = way:Find("highway")
	if highway ~= "" then
		way:Layer("highway", false)
		add_tag(way, "highway", "name", "smoothness", "trail_visibility", "sac_scale", "via_ferrata_scale", "surface", "tracktype")

		-- Iterate over relations to determine hiking-route tags
		local route_attr = {}
		while true do
			local rel = way:NextRelation()
			if not rel then break end

			-- read the type of route
			local route_type = way:FindInRelation("route")

			if route_type ~= "" then
				-- e.g. "hiking_route=yes"
				route_attr[route_type.."_route"] = "yes"

				-- Copy certain attributes from relations to the way
				for _, osm_attr_key in pairs(route_attributes) do
					local attr_key = route_type.."_"..osm_attr_key
					local osm_attr_value = way:FindInRelation(osm_attr_key)
					if osm_attr_value ~= nil and osm_attr_value ~= "" then
						local attr_value = route_attr[attr_key]
						if attr_value ~= nil and attr_value ~= "" then
							attr_value = attr_value..", "..osm_attr_value
						else
							attr_value = osm_attr_value
						end
						route_attr[attr_key] = attr_value
					end
				end
			end
		end

		-- Copy attributes from dictionary to the way object
		for attr, value in pairs(route_attr) do
			if value ~= "" then
				way:Attribute(attr, value)
			end
		end

		local label = way:Find("name")
		local hiking_route = route_attr["hiking_route"]
		local hiking_ref = route_attr["hiking_ref"]
		if hiking_route ~= nil and hiking_route ~= "" and hiking_ref ~= nil and hiking_ref ~= "" then
			if label ~= "" then
				label = label .. " (" .. hiking_ref .. ")"
			else
				label = hiking_ref
			end
		end
		way:Attribute("label", label)

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
		if natural == "cliff" or natural == "arete" then
			way:Layer("landuse", false)
		else
			way:Layer("landuse", true)
		end
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
		-- Only consider certain tags as lines. Everything else is considered to be a polygon, especially stations and platforms
		if any_of(railway, "rail", "narrow_gauge", "funicular", "light_rail", "monorail", "subway", "tram", "disused", "abandoned") then
			way:Layer("railway", false)
		else
			way:Layer("railway", true)
		end

		add_tag(way, "railway", "name", "tunnel")
		return
	end

	local aerialway = way:Find("aerialway")
	if aerialway ~= "" then
		way:Layer("aerialway", false)
		add_tag(way, "aerialway", "name", "access")
		return
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
	local rel_type = relation:Find("type")
	if rel_type == "boundary" or rel_type == "route" then
		relation:Accept()
	end
end
