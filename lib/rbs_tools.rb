# Utilities to extract method types information from inlined RBS signatures
# and convert them to JSON Schema properties
module RBSTools
  class MethodDecl < SimpleDelegator
    class Args < SimpleDelegator
      def required_keywords = super.transform_values { RBSTools.property_type_from_rbs(_1.type) }
      def optional_keywords = super.transform_values { RBSTools.property_type_from_rbs(_1.type) }
    end

    def description = @description ||= comment&.string&.gsub(/^@rbs.*$/, "")&.chomp
    def args = @args ||= Args.new(overloads.first.method_type.type)
  end

  # Add pattern matching support to RBS types
  using(Module.new do
    refine RBS::Types::ClassInstance do
      def deconstruct_keys(_) = {name:}
    end

    refine RBS::TypeName do
      def deconstruct_keys(_) = {kind:, name:}
    end
  end)
  extend self

  def property_type_from_rbs(rbs)
    case rbs
    in RBS::Types::Union
      enum = rbs.types.map(&:literal)
      type = enum.first.is_a?(Numeric) ? "number" : "string"
      {type: "string", enum:}
    in RBS::Types::ClassInstance[name: {kind: :class, name: :Date}]
      {type: "string", format: "date"}
    in RBS::Types::ClassInstance[name: {kind: :class, name: :Integer}]
      {type: "integer"}
    else
      {type: "string"}
    end
  end

  def method_sigs_for_class(class_name)
    # First, generate and collect RBS declarations for the class methods
    source_file_path = Object.const_source_location(class_name).first
    # Source file path could be virtual (eval) for anonymous classes
    return @openai_tools_schema = nil unless File.file?(source_file_path)

    # Get Prism AST
    ast = Prism.parse_file(source_file_path)
    # Extract RBS declarations from it using rbs-inline
    _, decls, _ = RBS::Inline::Parser.parse(ast, opt_in: false)
    writer = RBS::Inline::Writer.new
    rbs = []
    writer.translate_decl(decls.first, rbs)

    # Prepare RBS environment
    rbs_env = RBS::Environment.new
    rbs.each { rbs_env << _1 }

    # Generate RBS type for the class
    *path, name = *class_name.split("::").map(&:to_sym)
    namespace = path.empty? ? RBS::Namespace.root : RBS::Namespace.new(absolute: true, path:)
    class_rbs_type = RBS::TypeName.new(name:, namespace:)

    class_decl = rbs_env.class_decls[class_rbs_type]
    return @openai_tools_schema = nil unless class_decl

    # This is cryptic, but we cannot really load a proper RBS env, because most types would
    # be missing; so, we can rely only on the AST
    class_decl.decls.first.decl.members.select { RBS::AST::Members::MethodDefinition === _1 }
      .index_by(&:name)
      .transform_values do |decl|
        MethodDecl.new(decl)
      end
  end
end
