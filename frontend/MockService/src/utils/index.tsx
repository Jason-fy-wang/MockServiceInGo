export const IsResponse2xx = (status: number) => {
  return status >= 200 && status < 300;
};

export const IsResponse3xx = (status: number) => {
  return status >= 300 && status < 400;
};

export const IsResponse4xx = (status: number) => {
  return status >= 400 && status < 500;
};

export const IsResponse5xx = (status: number) => {
  return status >= 500 && status < 600;
};
