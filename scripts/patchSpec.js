#!/usr/bin/env node

function recursionProcess(response, prefix = '') {
  for (let key in response){
    if (key === "authorizationCode") {
      response[key].tokenUrl = "https://zoom.us/oauth/token"
      response[key].refreshUrl = "https://zoom.us/oauth/token"
    }

    if (response[key].type === 'array' && response[key].maximum) {
      response[key].maxItems = response[key].maximum;
      delete response[key].maximum;
    }

    if (response[key].type === 'array' && response[key].enum) {
      response[key].items.enum = response[key].enum
      delete response[key].enum
    }

    if (key === 'parameters') {
      response[key] = response[key].map((parameter) => {
        if (parameter.in === 'path') {
          parameter.required = true;
        }
        return parameter;
      });
    }

    if (key === 'paths') {
      response[key]['/phone/blocked_list/{blockedListId}'] = response[key]['/phone/blocked_list/{accountBlockedId}']
      delete response[key]['/phone/blocked_list/{accountBlockedId}']

      response[key]['/phone/call_queues/{callQueueId}/phone_numbers'] = response[key]['/phone/call_queues/{groupId}/phone_numbers']
      delete response[key]['/phone/call_queues/{groupId}/phone_numbers']

      response[key]['/phone/call_queues/{callQueueId}/phone_numbers/{phoneNumberId}'] = response[key]['/phone/call_queues/{groupId}/phone_numbers/{phoneNumberId}']
      delete response[key]['/phone/call_queues/{groupId}/phone_numbers/{phoneNumberId}']

      response[key]['/phone/shared_line_groups/{sharedLineGroupId}'].delete = response[key]['/phone/shared_line_groups/{slgId}'].delete
      delete response[key]['/phone/shared_line_groups/{slgId}']

      response[key]['/phone/shared_line_groups/{sharedLineGroupId}/members'] = response[key]['/phone/shared_line_groups/{slgId}/members']
      delete response[key]['/phone/shared_line_groups/{slgId}/members']

      response[key]['/phone/shared_line_groups/{sharedLineGroupId}/members/{memberId}'] = response[key]['/phone/shared_line_groups/{slgId}/members/{memberId}']
      delete response[key]['/phone/shared_line_groups/{slgId}/members/{memberId}']

      response[key]['/phone/shared_line_groups/{sharedLineGroupId}/phone_numbers'] = response[key]['/phone/shared_line_groups/{slgId}/phone_numbers']
      delete response[key]['/phone/shared_line_groups/{slgId}/phone_numbers']

      response[key]['/phone/shared_line_groups/{sharedLineGroupId}/phone_numbers/{phoneNumberId}'] = response[key]['/phone/shared_line_groups/{slgId}/phone_numbers/{phoneNumberId}']
      delete response[key]['/phone/shared_line_groups/{slgId}/phone_numbers/{phoneNumberId}']

      response[key]['/phone/users/{userId}/calling_plans/{type}'] = response[key]['/phone/users/{userId}/calling_plans/{planType}']
      delete response[key]['/phone/users/{userId}/calling_plans/{planType}']
    }

    if (key === '/phone/numbers/{phoneNumberId}') {
      response[key].patch.parameters = response[key].patch.parameters.map((parameter) => {
        if (parameter.in === 'path' && parameter.name === 'numberId') {
          parameter.name = 'phoneNumberId';
        }
        return parameter;
      });
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

const buffers = [];

(async () => {
  for await (const chunk of process.stdin) {
    buffers.push(chunk);
  }

  const buffer = Buffer.concat(buffers);
  const text = buffer.toString();
  const spec = JSON.parse(text);

  recursionProcess(spec);

  process.stdout.write(JSON.stringify(spec, null, 2));
})();
