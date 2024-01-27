-- Nodes will only be processed if one of these keys is present
node_keys = {
	"amenity",
	"label",
	"natural",
	"place",
	"railway",
	"shop",
	"waterway"
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
	local amenity = Find("amenity")
	local shop = Find("shop")
	local natural = Find("natural")
	local waterway = Find("waterway")
	if amenity ~= "" or shop ~= "" or natural ~= "" or waterway ~= "" then
		Layer("poi", false)
		add_tag(node, "name", "amenity", "shop", "natural", "ele", "waterway")
	end

	local label = Find("label")
	local category = Find("category")
	if label == "yes" then
		Layer("label", false)
		add_tag(node, "type", "category", "text")
		return
	end

	local place = Find("place")
	if place ~= "" then
		Layer("label", false)
		add_tag(node, "place", "name")
		return
	end

	local railway = Find("railway")
	if railway ~= "" then
		Layer("railway", false)
		add_tag(node, "railway", "name")
		return
	end
end

-- Assign ways to a layer, and set attributes, based on OSM tags
function way_function(way)
	local admin_level = tonumber(Find("admin_level")) or 1000
	local isBoundary = false
    local route_attr = {}

	-- Collect data from relation this way is part of
	while true do
		local relation = NextRelation()
		if not relation then break end

		-- read boundary information
		isBoundary = isBoundary or FindInRelation("boundary") == "administrative"
		admin_level = math.min(admin_level, tonumber(FindInRelation("admin_level")) or 1000)

        -- read the type of route
        local route_type = FindInRelation("route")

        if route_type ~= "" then
            -- e.g. "hiking_route=yes"
            route_attr[route_type.."_route"] = "yes"

            -- Copy certain attributes from relations to the way
            for _, osm_attr_key in pairs(route_attributes) do
                local attr_key = route_type.."_"..osm_attr_key
                local osm_attr_value = FindInRelation(osm_attr_key)
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

	-- Administrative boundaries in ways
	if isBoundary and admin_level ~= 1000 then
		Layer("boundary", false)
		add_tag(way, "boundary", "name", "protect_class", "border_type", "admin_level")
		Attribute("boundary", "administrative")
		Attribute("admin_level", admin_level)
	end

	local highway = Find("highway")
	if highway ~= "" then
		Layer("highway", false)
		add_tag(way, "highway", "name", "smoothness", "trail_visibility", "sac_scale", "via_ferrata_scale", "surface", "tracktype", "hiking_route", "hiking_ref")

		-- Copy attributes from dictionary to the way object
		for attr, value in pairs(route_attr) do
			if value ~= "" then
				Attribute(attr, value)
			end
		end

		local label = Find("name")
		local hiking_route = route_attr["hiking_route"]
		local hiking_ref = route_attr["hiking_ref"]
		if hiking_route ~= nil and hiking_route ~= "" and hiking_ref ~= nil and hiking_ref ~= "" then
			if label ~= "" then
				label = label .. " (" .. hiking_ref .. ")"
			else
				label = hiking_ref
			end
		end
		Attribute("label", label)

		return
	end

	local landuse = Find("landuse")
	if landuse ~= "" then
		Layer("landuse", true)
		add_tag(way, "landuse", "name")
		return
	end

	local natural = Find("natural")
	if natural ~= "" then
		-- Put "natural" things also to landuse for convenience
		if natural == "cliff" or natural == "arete" then
			Layer("landuse", false)
		else
			Layer("landuse", true)
		end
		add_tag(way, "name")
		Attribute("landuse", natural)
		return
	end

	local building = Find("building")
	if building ~= "" then
		Layer("building", true)
		return
	end

	local waterway = Find("waterway")
	if waterway ~= "" then
		Layer("waterway", false)
		add_tag(way, "waterway", "name", "intermittent", "tunnel")
		return
	end

	local railway = Find("railway")
	--if railway == "rail" or railway == "narrow_gauge" or railway == "funicular" or railway == "light_rail" or railway == "monorail" or railway == "subway" or railway == "tram" or railway == "disused" or railway == "abandoned" then
	if railway ~= "" then
		-- Only consider certain tags as lines. Everything else is considered to be a polygon, especially stations and platforms
		if any_of(railway, "rail", "narrow_gauge", "funicular", "light_rail", "monorail", "subway", "tram", "disused", "abandoned") then
			Layer("railway", false)
		else
			Layer("railway", true)
		end

		add_tag(way, "railway", "name", "tunnel")
		return
	end

	local aerialway = Find("aerialway")
	if aerialway ~= "" then
		Layer("aerialway", false)
		add_tag(way, "aerialway", "name", "access")
		return
	end
end

function relation_function(relation)
	local boundary = Find("boundary")

	-- administrative boundaries are handled on ways
	local isAdministrative = Find("admin_level") ~= ""

	if boundary ~= "" and not isAdministrative then
		Layer("boundary", false)
		add_tag(relation, "boundary", "name", "protect_class", "border_type", "admin_level")
		return
	end
end

function add_tag(feature, ...)
	local tags = {...}
	for i = 1, #tags do
		Attribute(tags[i], Find(tags[i]))
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
	local rel_type = Find("type")
	if rel_type == "boundary" or rel_type == "route" then
		Accept()
	end
end
