#!/usr/bin/env node

function snakeToCamel(str) {
  return str.replace(
    /(?!^)_(.)/g,
    (_, char) => char.toUpperCase()
  );
}

function upperFirst(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

function recursionProcess(response, prefix = '') {
  for (let key in response){
    if (key === "authorizationCode") {
      const url = "https://zoom.us/oauth/token";
      response[key].tokenUrl = url;
      response[key].refreshUrl = url;
    }

    if (response[key] !== void 0 && response[key].type === 'array' && response[key].maximum) {
      response[key].maxItems = response[key].maximum;
      delete response[key].maximum;
    }

    if (response[key] !== void 0 && response[key].type === 'array' && response[key].enum) {
      response[key].items.enum = response[key].enum
      delete response[key].enum
    }

    // The path parameters must always be required to be true
    if (key === 'parameters') {
      response[key] = response[key].map((parameter) => {
        if (parameter.in === 'path') {
          parameter.required = true;
        }
        return parameter;
      });
    }

    // The ogen does not support uniqueItems
    if (key === 'uniqueItems') {
      response[key] = false
    }

    // terraform doesn't have enum type
    if (key === 'enum') {
      delete response[key]
      continue
    }

    if (typeof response[key] !== "object") {
      continue
    }

    if(Array.isArray(response[key])) {
      response[key].forEach(function(item){
        recursionProcess(item, `${prefix} ${key}`);
      });
    } else {
      recursionProcess(response[key], `${prefix} ${key}`);
    }
  }
}

// Enable convenient errors for ogen
function enableConvenientErrorsPatch(spec) {
  // Since no response is defined for statuses above 400, use convenient errors to replace the default error response.
  spec.paths = Object.fromEntries(Object.entries(spec.paths).map(([path, pathValue]) => {
    return [path, Object.fromEntries(Object.entries(pathValue).map(([method, methodValue]) => {
      const filteredResponses = Object.fromEntries(Object.entries(methodValue.responses).filter(([responseCode, responseValue]) => {
        return Number(responseCode) < 400;
      }));

      methodValue.responses = {
        ...filteredResponses,
        default: {
          description: "For ogen convenient errors",
          content: {
            "application/json": {
              schema: {
                $ref: "#/components/schemas/ErrorResponse"
              },
            },
          },
        },
      };
      return [method, methodValue];
    }))];
  }));

  spec.components.schemas = {
    ...spec.components.schemas,
    ErrorResponse: {
      type: "object",
      properties: {
        code: {
          type: "integer",
        },
        message: {
          type: "string",
        },
      },
    },
  };
}

function phonePatch(spec) {
  /**
   * The oneOf atrribute of ogen is
   * - Unique per-object attributes
   * - primitive types
   * - discriminator attribute
   * only three patterns are supported.
   *
   * The following uses the discriminator attribute to support onfOf.
   */
  ["post", "patch"].forEach((method) => {
    const path = "/phone/extension/{extensionId}/call_handling/settings/{settingType}";
    if (!spec.paths[path]) {
      return;
    }

    const schemaObjectPrefix = `${upperFirst(method)}CallHandlingSettings`;

    const content = spec.paths[path][method].requestBody.content['application/json'];
    content.schema = {
      ...content.schema,
      discriminator: {
        propertyName: "sub_setting_type",
        mapping: Object.fromEntries(content.schema.oneOf.map((item) => {
          const objectName = upperFirst(snakeToCamel(item.properties.sub_setting_type.example));

          return [item.properties.sub_setting_type.example, `#/components/schemas/${schemaObjectPrefix}${objectName}`];
        })),
      },
    };

    spec.components.schemas = {
      ...spec.components.schemas,
      ...Object.fromEntries(content.schema.oneOf.map((item) => {
          const objectName = upperFirst(snakeToCamel(item.properties.sub_setting_type.example));
          return [`${schemaObjectPrefix}${objectName}`, item];
      })),
    };

    content.schema.oneOf = content.schema.oneOf.map((item) => {
      const objectName = upperFirst(snakeToCamel(item.properties.sub_setting_type.example));
      return {
        $ref: `#/components/schemas/${schemaObjectPrefix}${objectName}`,
      };
    });
  });

  // Some path names and path parameter values do not match,
  // so matching path parameters and path names.
  const replacePathsMappings = [
    {
      before: "/phone/blocked_list/{accountBlockedId}",
      after: "/phone/blocked_list/{blockedListId}",
    },
    {
      before: "/phone/call_queues/{groupId}/phone_numbers",
      after: "/phone/call_queues/{callQueueId}/phone_numbers",
    },
    {
      before: "/phone/call_queues/{groupId}/phone_numbers/{phoneNumberId}",
      after: "/phone/call_queues/{callQueueId}/phone_numbers/{phoneNumberId}",
    },
    {
      before: "/phone/shared_line_groups/{slgId}/members",
      after: "/phone/shared_line_groups/{sharedLineGroupId}/members",
    },
    {
      before: "/phone/shared_line_groups/{slgId}/members/{memberId}",
      after: "/phone/shared_line_groups/{sharedLineGroupId}/members/{memberId}",
    },
    {
      before: "/phone/shared_line_groups/{slgId}/phone_numbers",
      after: "/phone/shared_line_groups/{sharedLineGroupId}/phone_numbers",
    },
    {
      before: "/phone/shared_line_groups/{slgId}/phone_numbers/{phoneNumberId}",
      after: "/phone/shared_line_groups/{sharedLineGroupId}/phone_numbers/{phoneNumberId}",
    },
    {
      before: "/phone/users/{userId}/calling_plans/{planType}",
      after: "/phone/users/{userId}/calling_plans/{type}",
    },
  ]
  replacePathsMappings.forEach(({ before, after }) => {
    if (!spec.paths[before]) {
      return;
    }

    spec.paths[after] = spec.paths[before];
    delete spec.paths[before];
  });

  // Merging because the methods are separated in the same path.
  if (spec.paths['/phone/shared_line_groups/{slgId}']) {
    spec.paths['/phone/shared_line_groups/{sharedLineGroupId}'].delete = spec.paths['/phone/shared_line_groups/{slgId}'].delete
    delete spec.paths['/phone/shared_line_groups/{slgId}']
  }

  // The path name and parameter name do not match,
  // so the path parameter is matched to the path name.
  [
    '/phone/shared_line_groups/{sharedLineGroupId}/members',
    '/phone/shared_line_groups/{sharedLineGroupId}/phone_numbers',
    '/phone/shared_line_groups/{sharedLineGroupId}/phone_numbers/{phoneNumberId}',
  ].forEach((path) => {
    if (!spec.paths[path]) {
      return;
    }

    spec.paths[path] = Object.fromEntries(Object.entries(spec.paths[path]).map(([method, value]) => {
      value.parameters = value.parameters.map((parameter) => {
        if (parameter.in === 'path' && parameter.name === 'slgId') {
          parameter.name = 'sharedLineGroupId';
        }
        return parameter;
      });

      return [method, value];
    }));
  });

  // The path name and parameter name do not match,
  // so the path parameter is matched to the path name.
  [
    '/phone/numbers/{phoneNumberId}',
  ].forEach((path) => {
    if (!spec.paths[path]) {
      return;
    }

    spec.paths[path] = Object.fromEntries(Object.entries(spec.paths[path]).map(([method, value]) => {
      value.parameters = value.parameters.map((parameter) => {
        if (parameter.in === 'path' && parameter.name === 'numberId') {
          parameter.name = 'phoneNumberId';
        }
        return parameter;
      });

      return [method, value];
    }));
  });

  // POST /phone/call_queues doesn't require site_id required parameter
  if (spec.paths['/phone/call_queues']) {
    spec.paths['/phone/call_queues']['post']['requestBody']['content']['application/json']['schema']['required'] = ['name']
  }

  // GET /phone/call_queues/{callQueueId}/members doesn't have page_size and next_page_token parameters
  if (spec.paths['/phone/call_queues/{callQueueId}/members']) {
    spec.paths['/phone/call_queues/{callQueueId}/members']['get']['parameters'] =
        spec.paths['/phone/call_queues/{callQueueId}/members']['get']['parameters'].concat(
            { "name": "page_size", "in": "query", "required": false, "schema": { "type": "integer" } },
            { "name": "next_page_token", "in": "query", "required": false, "schema": { "type": "string" } },
        )
  }
}

const buffers = [];

(async () => {
  for await (const chunk of process.stdin) {
    buffers.push(chunk);
  }

  const buffer = Buffer.concat(buffers);
  const text = buffer.toString();
  const spec = JSON.parse(text);

  enableConvenientErrorsPatch(spec);
  phonePatch(spec);
  recursionProcess(spec);

  process.stdout.write(JSON.stringify(spec, null, 2));
})();
