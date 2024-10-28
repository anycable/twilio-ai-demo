class TodosController < ApplicationController
  before_action :set_todo, only: %i[ show edit update destroy ]

  # GET /todos
  def index
    @todos = Todo.all.public_send(params[:scope] || :current).then do
      _1.group_by(&:deadline)
    end
  end

  # GET /todos/new
  def new
    @todo = Todo.new
  end

  # GET /todos/1/edit
  def edit
  end

  # POST /todos
  def create
    @todo = Todo.new(todo_params)

    if @todo.save
      redirect_to todos_path
    else
      render :new, status: :unprocessable_entity
    end
  end

  # PATCH/PUT /todos/1
  def update
    if @todo.update(todo_params)
      if turbo_frame_request?
        render partial: "todo", locals: {todo: @todo}
      else
        redirect_back fallback_location: todos_path
      end
    else
      render :edit, status: :unprocessable_entity
    end
  end

  # DELETE /todos/1
  def destroy
    @todo.destroy!
    redirect_back fallback_location: todos_path
  end

  private
    # Use callbacks to share common setup or constraints between actions.
    def set_todo
      @todo = Todo.find(params.expect(:id))
    end

    # Only allow a list of trusted parameters through.
    def todo_params
      params.expect(todo: [:description, :completed, :deadline])
    end
end
