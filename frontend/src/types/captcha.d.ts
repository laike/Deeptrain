declare namespace grecaptcha {
  namespace enterprise {
    function render(id: string, options: {
      sitekey: string,
      size?: string,
      theme?: string,
      callback?: (response: string) => any,
    });
  }
}
