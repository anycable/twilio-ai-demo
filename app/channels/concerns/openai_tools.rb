module OpenAITools
  extend ActiveSupport::Concern

  class_methods do
    def tool(method_name)
      openai_tools << method_name
    end

    def openai_tools = @openai_tools ||= []

    def openai_tools_schema
      return @openai_tools_schema if defined?(@openai_tools_schema)

      return @openai_tools_schema = nil if openai_tools.empty?

      method_decls = RBSTools.method_sigs_for_class(self.name)

      # Generate tools configuration for the agents based on the registered tools
      openai_tools.inject([]) do |tools, method_name|
        next tools unless method_decls.key?(method_name)

        decl = method_decls[method_name]

        required = []
        properties = {}

        decl.args.required_keywords.each do |name, param|
          required << name.to_s
          properties[name.to_s] = param
        end

        decl.args.optional_keywords.each do |name, param|
          properties[name.to_s] = param
        end

        parameters = {
          type: "object",
          properties:,
          required:
        }

        # Get description from the method comment
        description = decl.description

        spec = {
          type: "function",
          name: method_name,
          description:,
          parameters:
        }

        tools << spec
        tools
      end
    end
  end
end
